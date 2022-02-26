// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package transfers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/moov-io/ach"

	"github.com/moov-io/paygate/pkg/client"
)

type Repository interface {
	getTransfers(orgID string, params transferFilterParams) ([]*client.Transfer, error)
	GetTransfer(id string) (*client.Transfer, error)
	UpdateTransferStatus(transferID string, status client.TransferStatus) error
	WriteUserTransfer(orgID string, transfer *client.Transfer) error
	deleteUserTransfer(orgID string, transferID string) error

	SaveReturnCode(transferID string, returnCode string) error
	saveTraceNumbers(transferID string, traceNumbers []string) error
	getTraceNumbers(transferID string) ([]string, error)

	LookupTransferFromReturn(amount client.Amount, traceNumber string, effectiveEntryDate time.Time) (*client.Transfer, error)
}

func NewRepo(db *sql.DB) *sqlRepo {
	return &sqlRepo{db: db}
}

type sqlRepo struct {
	db *sql.DB
}

func (r *sqlRepo) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *sqlRepo) getTransfers(orgID string, params transferFilterParams) ([]*client.Transfer, error) {
	var query strings.Builder
	query.WriteString("select transfer_id from transfers where ")

	var args []interface{}
	query.WriteString("organization = ? and created_at >= ? and created_at <= ? and deleted_at is null ")
	args = append(args, orgID, params.StartDate, params.EndDate)

	if string(params.Status) != "" {
		query.WriteString("and status = ? ")
		args = append(args, params.Status)
	}

	if len(params.CustomerIDs) > 0 {
		s := fmt.Sprintf(
			"and ( source_customer_id in (?%[1]s) or destination_customer_id in (?%[1]s) ) ",
			strings.Repeat(",?", len(params.CustomerIDs)-1),
		)
		query.WriteString(s)
		for i := 0; i < len(params.CustomerIDs)*2; i++ {
			args = append(args, params.CustomerIDs[i%len(params.CustomerIDs)])
		}
	}

	query.WriteString("order by created_at desc limit ? offset ?;")
	args = append(args, params.Count, params.Skip)

	stmt, err := r.db.Prepare(query.String())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transferIDs []string
	transfers := make([]*client.Transfer, 0) // allocate array so JSON marshal is [] instead of null

	for rows.Next() {
		var row string
		if err := rows.Scan(&row); err != nil {
			return transfers, fmt.Errorf("getTransfers scan: %v", err)
		}
		if row != "" {
			transferIDs = append(transferIDs, row)
		}
	}
	if err := rows.Err(); err != nil {
		return transfers, fmt.Errorf("getTransfers: rows.Err=%v", err)
	}

	// read each transferID
	for i := range transferIDs {
		t, err := r.getUserTransfer(transferIDs[i], orgID)
		if err == nil && t.TransferID != "" {
			transfers = append(transfers, t)
		}
	}
	return transfers, rows.Err()
}

func (r *sqlRepo) getUserTransfer(transferID string, orgID string) (*client.Transfer, error) {
	query := `select transfer_id, amount_currency, amount_value, source_customer_id, source_account_id, destination_customer_id, destination_account_id, description, status, same_day, return_code, processed_at, created_at
from transfers
where transfer_id = ? and organization = ? and deleted_at is null
limit 1`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var returnCode *string
	transfer := &client.Transfer{}

	err = stmt.QueryRow(transferID, orgID).Scan(
		&transfer.TransferID,
		&transfer.Amount.Currency,
		&transfer.Amount.Value,
		&transfer.Source.CustomerID,
		&transfer.Source.AccountID,
		&transfer.Destination.CustomerID,
		&transfer.Destination.AccountID,
		&transfer.Description,
		&transfer.Status,
		&transfer.SameDay,
		&returnCode,
		&transfer.ProcessedAt,
		&transfer.Created,
	)
	if transfer.TransferID == "" || err != nil {
		return nil, err
	}

	// query the trace table
	// append the transfer if any tracenums
	traceNumbers, err := r.getTraceNumbers(transferID)
	if err != nil {
		return nil, err
	}
	for i := range traceNumbers {
		transfer.TraceNumbers = append(transfer.TraceNumbers, traceNumbers[i])
	}
	if returnCode != nil {
		if rc := ach.LookupReturnCode(*returnCode); rc != nil {
			transfer.ReturnCode = &client.ReturnCode{
				Code:        rc.Code,
				Reason:      rc.Reason,
				Description: rc.Description,
			}
		}
	}
	return transfer, nil
}

func (r *sqlRepo) GetTransfer(transferID string) (*client.Transfer, error) {
	query := `select organization from transfers where transfer_id = ? and deleted_at is null limit 1`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	orgID := ""
	if err := stmt.QueryRow(transferID).Scan(&orgID); err != nil {
		return nil, err
	}
	return r.getUserTransfer(transferID, orgID)
}

func (r *sqlRepo) UpdateTransferStatus(transferID string, status client.TransferStatus) error {
	query := `update transfers set status = ? where transfer_id = ? and deleted_at is null`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, transferID)
	return err
}

func (r *sqlRepo) WriteUserTransfer(orgID string, transfer *client.Transfer) error {
	query := `insert into transfers (transfer_id, organization, amount_currency, amount_value, source_customer_id, source_account_id, destination_customer_id, destination_account_id, description, status, same_day, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		transfer.TransferID,
		orgID,
		transfer.Amount.Currency,
		transfer.Amount.Value,
		transfer.Source.CustomerID,
		transfer.Source.AccountID,
		transfer.Destination.CustomerID,
		transfer.Destination.AccountID,
		transfer.Description,
		transfer.Status,
		transfer.SameDay,
		time.Now(),
	)
	return err
}

func (r *sqlRepo) deleteUserTransfer(orgID string, transferID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `select status from transfers where transfer_id = ? and organization = ? and deleted_at is null limit 1;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	var status string
	if err := stmt.QueryRow(transferID, orgID).Scan(&status); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if !strings.EqualFold(status, string(client.PENDING)) {
		tx.Rollback()
		return fmt.Errorf("transferID=%s is not in PENDING status", transferID)
	}

	query = `update transfers set deleted_at = ?
where transfer_id = ? and organization = ? and status = ? and deleted_at is null`
	stmt, err = tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), transferID, orgID, client.PENDING)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	return tx.Commit()
}

func (r *sqlRepo) SaveReturnCode(transferID string, returnCode string) error {
	query := `update transfers set return_code = ? where transfer_id = ? and return_code is null and deleted_at is null`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(returnCode, transferID)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (r *sqlRepo) saveTraceNumbers(transferID string, traceNumbers []string) error {
	query := `insert into transfer_trace_numbers(transfer_id, trace_number) values (?, ?);`
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	for i := range traceNumbers {
		if _, err := stmt.Exec(transferID, traceNumbers[i]); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *sqlRepo) LookupTransferFromReturn(amount client.Amount, traceNumber string, effectiveEntryDate time.Time) (*client.Transfer, error) {
	// To match returned files we take a few values which are assumed to uniquely identify a Transfer.
	// traceNumber, per NACHA guidelines, should be globally unique (routing number + random value),
	// but we are going to filter to only select Transfers created within a few days of the EffectiveEntryDate
	// to avoid updating really old (or future, I suppose) objects.
	query := `select xf.transfer_id, xf.organization from transfers as xf
inner join transfer_trace_numbers trace on xf.transfer_id = trace.transfer_id
where xf.amount_value = ? and trace.trace_number = ? and xf.status = ? and (xf.created_at > ? and xf.created_at < ?) and xf.deleted_at is null limit 1`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	transferId, orgID := "", ""
	min, max := startOfDayAndTomorrow(effectiveEntryDate)
	// Only include Transfer objects within 5 calendar days of the EffectiveEntryDate
	min = min.Add(-5 * 24 * time.Hour)
	max = max.Add(5 * 24 * time.Hour)

	row := stmt.QueryRow(amount.Value, traceNumber, client.PROCESSED, min, max)
	if err := row.Scan(&transferId, &orgID); err != nil {
		return nil, err
	}

	return r.getUserTransfer(transferId, orgID)
}

// startOfDayAndTomorrow returns two time.Time values from a given time.Time value.
// The first is at the start of the same day as provided and the second is exactly 24 hours
// after the first.
func startOfDayAndTomorrow(in time.Time) (time.Time, time.Time) {
	start := in.Truncate(24 * time.Hour)
	return start, start.Add(24 * time.Hour)
}

func (r *sqlRepo) getTraceNumbers(transferID string) ([]string, error) {
	var traceNumbers []string
	query := `select trace_number from transfer_trace_numbers
where transfer_id = ?`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	args := []interface{}{transferID}
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row string
		if err := rows.Scan(&row); err != nil {
			return traceNumbers, fmt.Errorf("getTraceNumbers scan: %v", err)
		}
		if row != "" {
			traceNumbers = append(traceNumbers, row)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return traceNumbers, nil
}
