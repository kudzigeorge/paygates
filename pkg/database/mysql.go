// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base/docker"

	kitprom "github.com/go-kit/kit/metrics/prometheus"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/lopezator/migrator"
	"github.com/moov-io/base/log"
	"github.com/ory/dockertest/v3"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

var (
	mysqlConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "mysql_connections",
		Help: "How many MySQL connections and what status they're in.",
	}, []string{"state"})

	// mySQLErrDuplicateKey is the error code for duplicate entries
	// https://dev.mysql.com/doc/refman/8.0/en/server-error-reference.html#error_er_dup_entry
	mySQLErrDuplicateKey uint16 = 1062

	maxActiveMySQLConnections = func() int {
		if v := os.Getenv("MYSQL_MAX_CONNECTIONS"); v != "" {
			if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
				return int(n)
			}
		}
		return 16
	}()

	mysqlMigrations = migrator.Migrations(
		execsql(
			"create_namespace_configs",
			`create table namespace_configs(namespace varchar(40) primary key not null, company_identification varchar(40) not null)`,
		),
		execsql(
			"create_transfers",
			`create table if not exists transfers(transfer_id varchar(40) primary key not null, namespace varchar(40) not null, amount_currency varchar(3) not null, amount_value integer not null, source_customer_id varchar(40) not null, source_account_id varchar(40) not null, destination_customer_id varchar(40) not null, destination_account_id varchar(40) not null, description varchar(200) not null, status varchar(10) not null, same_day boolean not null, return_code varchar(10), created_at datetime not null, last_updated_at datetime not null, deleted_at datetime);`,
		),
		execsql(
			"add_remote_addr_to_transfers",
			// Max length for IPv6 addresses -- https://stackoverflow.com/a/7477384
			"alter table transfers add column remote_address varchar(45) not null default '';",
		),
		execsql(
			"add_micro_deposits",
			"create table micro_deposits(micro_deposit_id varchar(40) primary key not null, destination_customer_id varchar(40) not null, destination_account_id varchar(40) not null, status varchar(10) not null, created_at datetime not null, deleted_at datetime);",
		),
		execsql(
			"create_micro_deposits__account_id_idx",
			`create unique index micro_deposits_account_id on micro_deposits (destination_account_id);`,
		),
		execsql(
			"add_micro_deposit_amounts",
			"create table micro_deposit_amounts(micro_deposit_id varchar(40) not null, amount_currency varchar(3) not null, amount_value integer not null);",
		),
		execsql(
			"create_micro_deposit_amounts__account_id_idx",
			`create index micro_deposit_amounts_idx on micro_deposit_amounts (micro_deposit_id);`,
		),
		execsql(
			"create_micro_deposit_transfers",
			`create table micro_deposit_transfers(micro_deposit_id varchar(40) not null, transfer_id varchar(40) primary key not null);`,
		),
		execsql(
			"create_transfer_trace_numbers",
			`create table transfer_trace_numbers(transfer_id varchar(40) not null, trace_number varchar(20) not null);`,
		),
		execsql(
			"create_transfer_trace_numbers_unique_idx",
			`create unique index transfer_trace_numbers_idx on transfer_trace_numbers (transfer_id, trace_number);`,
		),
		execsql(
			"add_processed_at__to__transfers",
			`alter table transfers add column processed_at datetime;`,
		),
		execsql(
			"add_processed_at__to__micro_deposits",
			`alter table micro_deposits add column processed_at datetime;`,
		),
		execsql(
			"rename_namespace_configs_to_organization_configs",
			`alter table namespace_configs rename to organization_configs;`,
		),
		execsql(
			"rename_organization_configs_namespace_to_organization",
			`alter table organization_configs rename column namespace to organization;`,
		),
		execsql(
			"rename_transfers_namespace_to_organization",
			`alter table transfers rename column namespace to organization;`,
		),
	)
)

type discardLogger struct{}

func (l discardLogger) Print(v ...interface{}) {}

func init() {
	gomysql.SetLogger(discardLogger{})
}

type mysql struct {
	dsn    string
	logger log.Logger

	connections *kitprom.Gauge
}

func (my *mysql) Connect(ctx context.Context) (*sql.DB, error) {
	if my == nil {
		return nil, fmt.Errorf("nil %T", my)
	}

	db, err := sql.Open("mysql", my.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxActiveMySQLConnections)

	// Check out DB is up and working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Migrate our database
	if m, err := migrator.New(mysqlMigrations); err != nil {
		return nil, err
	} else {
		if err := m.Migrate(db); err != nil {
			return nil, err
		}
	}

	// Setup metrics after the database is setup
	go func() {
		t := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				stats := db.Stats()
				my.connections.With("state", "idle").Set(float64(stats.Idle))
				my.connections.With("state", "inuse").Set(float64(stats.InUse))
				my.connections.With("state", "open").Set(float64(stats.OpenConnections))
			}
		}
	}()

	return db, nil
}

func mysqlConnection(logger log.Logger, user, pass string, address string, database string) *mysql {
	timeout := "30s"
	if v := os.Getenv("MYSQL_TIMEOUT"); v != "" {
		timeout = v
	}
	params := fmt.Sprintf("timeout=%s&charset=utf8mb4&parseTime=true&sql_mode=ALLOW_INVALID_DATES", timeout)
	dsn := fmt.Sprintf("%s:%s@%s/%s?%s", user, pass, address, database, params)
	return &mysql{
		dsn:         dsn,
		logger:      logger,
		connections: mysqlConnections,
	}
}

// TestMySQLDB is a wrapper around sql.DB for MySQL connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestMySQLDB struct {
	DB *sql.DB

	container *dockertest.Resource

	shutdown func() // context shutdown func
}

func (r *TestMySQLDB) Close() error {
	r.shutdown()

	// Verify all connections are closed before closing DB
	if conns := r.DB.Stats().OpenConnections; conns != 0 {
		panic(fmt.Sprintf("found %d open MySQL connections", conns))
	}

	r.container.Close()

	return r.DB.Close()
}

// CreateTestMySQLDB returns a TestMySQLDB which can be used in tests
// as a clean mysql database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestMySQLDB.
func CreateTestMySQLDB(t *testing.T) *TestMySQLDB {
	if testing.Short() {
		t.Skip("-short flag enabled")
	}
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			"MYSQL_USER=moov",
			"MYSQL_PASSWORD=secret",
			"MYSQL_ROOT_PASSWORD=secret",
			"MYSQL_DATABASE=paygate",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pool.Retry(func() error {
		db, err := sql.Open("mysql", fmt.Sprintf("moov:secret@tcp(localhost:%s)/paygate", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	})
	if err != nil {
		resource.Close()
		t.Fatal(err)
	}

	logger := log.NewNopLogger()
	address := fmt.Sprintf("tcp(localhost:%s)", resource.GetPort("3306/tcp"))

	ctx, cancelFunc := context.WithCancel(context.Background())

	db, err := mysqlConnection(logger, "moov", "secret", address, "paygate").Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	t.Cleanup(func() {
		pool.Purge(resource)
	})

	return &TestMySQLDB{DB: db, container: resource, shutdown: cancelFunc}
}

// MySQLUniqueViolation returns true when the provided error matches the MySQL code
// for duplicate entries (violating a unique table constraint).
func MySQLUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), fmt.Sprintf("Error %d: Duplicate entry", mySQLErrDuplicateKey))
	if e, ok := err.(*gomysql.MySQLError); ok {
		return match || e.Number == mySQLErrDuplicateKey
	}
	return match
}
