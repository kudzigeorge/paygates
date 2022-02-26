// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fundflow

import (
	"fmt"
	"sort"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/base"
	"github.com/moov-io/base/stime"
	customers "github.com/moov-io/customers/pkg/client"

	"github.com/moov-io/paygate/pkg/achx"
	"github.com/moov-io/paygate/pkg/client"
	"github.com/moov-io/paygate/pkg/config"

	"github.com/moov-io/base/log"
)

// FirstPerson returns a Strategy for fund flows where PayGate runs as an ACH originator
// at an FI. This implies funds move in one direction from the FI -- either in or out.
//
// Outgoing credits are debited from the account at the FI without delay and the credits
// are posted after the RDFI receives the file.
//
// Debiting the remote account means we'll credit our account, but typically hold
// those funds for a settlement period.
//
// These transfers involve one file with an optional return from the RDFI which should trigger
// a reversal in the accounting ledger.
type FirstParty struct {
	cfg         config.ODFI
	logger      log.Logger
	timeService stime.TimeService
}

func NewFirstPerson(logger log.Logger, cfg config.ODFI) Strategy {
	return &FirstParty{
		cfg:         cfg,
		logger:      logger,
		timeService: stime.NewSystemTimeService(),
	}
}

func (fp *FirstParty) Originate(companyID string, xfer *client.Transfer, src Source, dst Destination) ([]*ach.File, error) {
	if src.Account.RoutingNumber == dst.Account.RoutingNumber {
		// Reject transfers that are within our ODFI. These should be internal to the ledger rather than
		// requiring an ACH file sent anywhere.
		return nil, fmt.Errorf("rejecting transfer between two accounts within %s", src.Account.RoutingNumber)
	}
	if src.Account.RoutingNumber != fp.cfg.RoutingNumber && dst.Account.RoutingNumber != fp.cfg.RoutingNumber {
		// First-Party transfers need to contain the ODFI as either the source or destination
		return nil, fmt.Errorf("rejecting third-party transfer between FI's we don't represent (source: %s, destination: %s)", src.Account.RoutingNumber, dst.Account.RoutingNumber)
	}
	source := achx.Source{
		Customer:      src.Customer,
		Account:       src.Account,
		AccountNumber: src.AccountNumber,
	}
	destination := achx.Destination{
		Customer:      dst.Customer,
		Account:       dst.Account,
		AccountNumber: dst.AccountNumber,
	}

	// If we are debiting the source that Customer's status needs to be VERIFIED
	if fp.cfg.RoutingNumber == destination.Account.RoutingNumber {
		if !strings.EqualFold(string(src.Customer.Status), string(customers.CUSTOMERSTATUS_VERIFIED)) {
			return nil, fmt.Errorf("source customerID=%s does not support debit with status %s", src.Customer.CustomerID, src.Customer.Status)
		}
	}

	opts := achx.Options{
		ODFIRoutingNumber:     fp.cfg.RoutingNumber,
		Gateway:               fp.cfg.Gateway,
		FileConfig:            fp.cfg.FileConfig,
		CutoffTimezone:        fp.cfg.Cutoffs.Location(),
		EffectiveEntryDate:    calculateEffectiveEntryDate(fp.cfg, fp.timeService, xfer.SameDay),
		CompanyIdentification: companyID,
	}
	// Balance entries from transfers which appear to not be "account validation" (aka micro-deposits).
	// Right now we're doing this by checking the amount which obviously isn't ideal.
	//
	// TODO(adam): Better detection for when to offset or not.
	opts.FileConfig.BalanceEntries = fp.cfg.FileConfig.BalanceEntries && (xfer.Amount.Value >= 50)

	file, err := achx.ConstructFile(xfer.TransferID, opts, xfer, source, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: transferID=%s: %v", xfer.TransferID, err)
	}
	return []*ach.File{file}, err
}

func (fp *FirstParty) HandleReturn(returned *ach.File, xfer *client.Transfer) ([]*ach.File, error) {
	return nil, nil
}

func calculateEffectiveEntryDate(cfg config.ODFI, ss stime.TimeService, sameDay bool) base.Time {
	when := base.NewTime(ss.Now().In(cfg.Cutoffs.Location()))
	afterCutoffs := afterCutoffWindows(cfg.Cutoffs, when)

	// If we're after-hours then handle the transfer's settlement for later on
	if afterCutoffs {
		if sameDay {
			return when.AddBankingDay(1)
		}
		return when.AddBankingDay(2)
	}

	// Handle transfers that are going out today still
	if sameDay {
		return when
	}

	return when.AddBankingDay(1)
}

func afterCutoffWindows(cfg config.Cutoffs, when base.Time) bool {
	if len(cfg.Windows) == 0 {
		return false
	}

	windows := make([]string, len(cfg.Windows))
	copy(windows, cfg.Windows)
	sort.Strings(windows)

	if len(windows) == 0 {
		return false
	}

	now := when.Format("15:04")
	return now >= windows[len(windows)-1]
}
