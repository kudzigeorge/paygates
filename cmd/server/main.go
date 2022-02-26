// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/moov-io/base/admin"
	"github.com/moov-io/base/log"
	"github.com/moov-io/paygate"
	"github.com/moov-io/paygate/pkg/config"
	configadmin "github.com/moov-io/paygate/pkg/config/admin"
	"github.com/moov-io/paygate/pkg/customers"
	"github.com/moov-io/paygate/pkg/customers/accounts"
	"github.com/moov-io/paygate/pkg/database"
	"github.com/moov-io/paygate/pkg/organization"
	"github.com/moov-io/paygate/pkg/transfers"
	transferadmin "github.com/moov-io/paygate/pkg/transfers/admin"
	"github.com/moov-io/paygate/pkg/transfers/fundflow"
	"github.com/moov-io/paygate/pkg/transfers/inbound"
	"github.com/moov-io/paygate/pkg/transfers/pipeline"
	"github.com/moov-io/paygate/pkg/upload"
	"github.com/moov-io/paygate/pkg/util"
	"github.com/moov-io/paygate/pkg/validation/microdeposits"
	"github.com/moov-io/paygate/x/route"
	"github.com/moov-io/paygate/x/schedule"
	"github.com/moov-io/paygate/x/trace"

	"github.com/gorilla/mux"
)

var (
	flagConfigFile = flag.String("config", "", "Filepath for config file to load")
)

func main() {
	flag.Parse()

	// Read our config file
	cfg := readConfig(os.Getenv("CONFIG_FILE"))
	cfg.Logger = cfg.Logger.Set("package", log.String("main"))

	_, traceCloser, err := trace.NewConstantTracer(cfg.Logger, "paygate")
	if err != nil {
		panic(fmt.Sprintf("ERROR starting tracer: %v", err))
	}
	defer traceCloser.Close()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// migrate database
	db, err := database.New(ctx, cfg.Logger, cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("error creating database: %v", err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			cfg.Logger.LogErrorf("exit: %v", err)
		}
	}()

	// Listen for application termination.
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Spin up admin HTTP server
	adminServer := admin.NewServer(cfg.Admin.BindAddress)
	adminServer.AddVersionHandler(paygate.Version) // Setup 'GET /version'
	go func() {
		cfg.Logger.Logf("admin: listening on %s", adminServer.BindAddr())
		if err := adminServer.Listen(); err != nil {
			err = cfg.Logger.LogErrorf("problem starting admin http: %v", err).Err()
			errs <- err
		}
	}()

	// Create HTTP handler
	handler := mux.NewRouter()
	route.PingRoute(cfg.Logger, handler)

	defer adminServer.Shutdown()

	// Register admin route for config marshaling
	configadmin.RegisterRoutes(adminServer, cfg)

	// Find our fundflow strategy
	fundflowStrategy := fundflow.NewFirstPerson(cfg.Logger, cfg.ODFI)

	// Setup our transfer publisher
	transferPublisher, err := pipeline.NewPublisher(cfg.Pipeline)
	if err != nil {
		panic(fmt.Sprintf("ERROR setting up transfer publisher: %v", err))
	}
	defer transferPublisher.Shutdown(ctx)

	transferSubscription, err := pipeline.NewSubscription(cfg)
	if err != nil {
		panic(fmt.Sprintf("ERROR setting up transfer subscription: %v", err))
	}
	defer transferSubscription.Shutdown(ctx)

	agent, err := upload.New(cfg.Logger, cfg.ODFI)
	if err != nil {
		// We don't want to crash the system on this failure. It's an important
		// connection, but not strictly required as the issue may be resolved
		// without a restart of PayGate.
		cfg.Logger.LogErrorf("problem with upload.Agent connection: %v", err)
	}
	defer agent.Close()
	adminServer.AddLivenessCheck(upload.Type(cfg.ODFI), agent.Ping)

	merger, err := pipeline.NewMerging(cfg.Logger, cfg.Pipeline)
	if err != nil {
		panic(fmt.Sprintf("ERROR setting up xfer merging: %v", err))
	}

	cutoffs, err := schedule.ForCutoffTimes(cfg.ODFI.Cutoffs.Timezone, cfg.ODFI.Cutoffs.Windows)
	if err != nil {
		panic(fmt.Sprintf("ERROR setting up cutoff times: %v", err))
	} else {
		cfg.Logger.Logf("registered %s cutoffs=%v", cfg.ODFI.Cutoffs.Timezone, strings.Join(cfg.ODFI.Cutoffs.Windows, ","))
	}

	pipelineRepo := pipeline.NewRepo(db)
	xferAgg, err := pipeline.NewAggregator(cfg, agent, pipelineRepo, merger, transferSubscription, nil)
	if err != nil {
		panic(fmt.Sprintf("ERROR creating transfer aggregator: %v", err))
	}
	defer xferAgg.Shutdown()
	go xferAgg.Start(ctx, cutoffs)
	xferAgg.RegisterRoutes(adminServer)

	// Customers
	customersClient := customers.NewClient(cfg.Logger, cfg.Customers, customers.HttpClient)
	adminServer.AddLivenessCheck("customers", customersClient.Ping)

	// Setup
	registerMicroDepositHealth(cfg, customersClient, adminServer)

	// Organization
	orgRepo := organization.NewRepo(db)
	organization.NewRouter(orgRepo).RegisterRoutes(handler)

	// Accounts
	accountDecryptor, err := accounts.NewDecryptor(cfg.Customers.Accounts.Decryptor, customersClient)
	if err != nil {
		panic(fmt.Sprintf("ERROR ccreating account decryptor: %v", err))
	}

	// Transfers
	transfersRepo := transfers.NewRepo(db)
	defer transfersRepo.Close()
	transfers.NewRouter(cfg, transfersRepo, orgRepo, customersClient, accountDecryptor, fundflowStrategy, transferPublisher).RegisterRoutes(handler)
	transferadmin.RegisterRoutes(cfg, adminServer, transfersRepo)

	// Micro-Deposit Validation
	microDepositRepo := microdeposits.NewRepo(db)
	microdeposits.NewRouter(cfg, microDepositRepo, transfersRepo, customersClient, accountDecryptor, fundflowStrategy, transferPublisher).RegisterRoutes(handler)

	// Create main HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Default port if not specified
	}
	port = ":" + port
	serve := &http.Server{
		Addr:    port,
		Handler: handler,
		TLSConfig: &tls.Config{
			InsecureSkipVerify:       false,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	shutdownServer := func() {
		if err := serve.Shutdown(context.TODO()); err != nil {
			cfg.Logger.LogErrorf("shutdown: %v", err)
		}
	}
	defer shutdownServer()

	// Start main HTTP server
	go func() {
		if certFile, keyFile := os.Getenv("HTTPS_CERT_FILE"), os.Getenv("HTTPS_KEY_FILE"); certFile != "" && keyFile != "" {
			cfg.Logger.Logf("startup: binding to %s for secure HTTP server", cfg.Http.BindAddress)
			if err := serve.ListenAndServeTLS(certFile, keyFile); err != nil {
				cfg.Logger.LogErrorf("exit: %v", err)
			}
		} else {
			cfg.Logger.Logf("startup: binding to %s for HTTP server", cfg.Http.BindAddress)
			if err := serve.ListenAndServe(); err != nil {
				cfg.Logger.LogErrorf("exit: %v", err)
			}
		}
	}()

	// Setup our inbound file processor and scheduler
	fileProcessors := inbound.SetupProcessors(
		inbound.NewCorrectionProcessor(cfg.Logger),
		inbound.NewPrenoteProcessor(cfg.Logger),
		inbound.NewReturnProcessor(cfg.Logger, transfersRepo),
	)
	inboundProcessor := inbound.NewPeriodicScheduler(cfg, agent, fileProcessors)
	go func() {
		if err := inboundProcessor.Start(); err != nil {
			panic(fmt.Sprintf("ERROR with inbound processor: %v", err))
		}
	}()
	defer inboundProcessor.Shutdown()

	if err := <-errs; err != nil {
		cfg.Logger.LogErrorf("exit: %v", err)
	}
}

var (
	exampleConfigFilepath = filepath.Join("examples", "config.yaml")
)

func readConfig(path string) *config.Config {
	path = util.Or(path, *flagConfigFile, exampleConfigFilepath)
	cfg, err := config.FromFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	cfg.Logger.Logf("starting paygate server version %s", paygate.Version)
	if err := validateTemplate(cfg.ODFI); err != nil {
		panic(fmt.Sprintf("ERROR %v", err))
	}
	return cfg
}

func validateTemplate(cfg config.ODFI) error {
	data := upload.FilenameData{
		RoutingNumber: cfg.RoutingNumber,
	}
	filename, err := upload.RenderACHFilename(cfg.FilenameTemplate(), data)
	if err != nil {
		return fmt.Errorf("invalid filename template: %v", err)
	}
	if filename == "" {
		return errors.New("empty filename rendered")
	}
	return nil
}

func registerMicroDepositHealth(cfg *config.Config, client customers.Client, svc *admin.Server) {
	if micro := cfg.Validation.MicroDeposits; micro != nil {
		check := customers.HealthChecker(client, micro.Source.Organization, micro.Source.CustomerID, micro.Source.AccountID)
		svc.AddLivenessCheck("micro-deposits-account", check)
	}
}
