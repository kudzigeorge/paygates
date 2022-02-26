// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fundflow

import (
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base/log"
	"github.com/moov-io/base/stime"
	customers "github.com/moov-io/customers/pkg/client"

	"github.com/moov-io/paygate/pkg/client"
	"github.com/moov-io/paygate/pkg/config"
)

func TestOriginate__DebitCheck(t *testing.T) {
	cfg := config.Empty()
	cfg.ODFI.RoutingNumber = "987654320"

	fp := NewFirstPerson(cfg.Logger, cfg.ODFI)

	companyID := "MOOV"
	xfer := &client.Transfer{}
	src := Source{
		Customer: customers.Customer{
			Status: customers.CUSTOMERSTATUS_RECEIVE_ONLY,
		},
	}
	dest := Destination{
		Account: customers.Account{
			RoutingNumber: "987654320",
		},
	}
	if _, err := fp.Originate(companyID, xfer, src, dest); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "does not support debit") {
			t.Error(err)
		}
	}
}

func TestOriginate__RoutingNumberErr(t *testing.T) {
	cfg := config.Empty() // leave off RoutingNumber for first test
	fp := NewFirstPerson(log.NewNopLogger(), cfg.ODFI)

	src := Source{
		Account: customers.Account{
			RoutingNumber: "987654320",
		},
	}
	dst := Destination{
		Account: customers.Account{
			RoutingNumber: "987654320",
		},
	}

	if file, err := fp.Originate("companyID", nil, src, dst); file != nil || err == nil {
		t.Fatal("expected nil File and error")
	} else {
		if !strings.Contains(err.Error(), "rejecting transfer between two accounts within") {
			t.Error(err)
		}
	}

	cfg.ODFI.RoutingNumber = "987654320"
	src.Account.RoutingNumber = "123"
	dst.Account.RoutingNumber = "987"

	if file, err := fp.Originate("companyID", nil, src, dst); file != nil || err == nil {
		t.Fatal("expected nil File and error")
	} else {
		if !strings.Contains(err.Error(), "rejecting third-party transfer between FI's we don't represent") {
			t.Error(err)
		}
	}
}

func TestOriginateFull(t *testing.T) {
	cfg := config.Empty()
	cfg.ODFI.RoutingNumber = "987654320"

	fp := NewFirstPerson(cfg.Logger, cfg.ODFI)

	companyID := "MOOV"
	xfer := &client.Transfer{
		Amount: client.Amount{
			Currency: "USD",
			Value:    153,
		},
		Description: "test payment",
	}
	src := Source{
		Customer: customers.Customer{
			Status: customers.CUSTOMERSTATUS_VERIFIED,
		},
		Account: customers.Account{
			Type:          customers.ACCOUNTTYPE_SAVINGS,
			RoutingNumber: "123456780",
		},
		AccountNumber: "123456",
	}
	dest := Destination{
		Account: customers.Account{
			Type:          customers.ACCOUNTTYPE_SAVINGS,
			RoutingNumber: "987654320",
		},
		AccountNumber: "654321",
	}

	files, err := fp.Originate(companyID, xfer, src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 1 {
		if err := files[0].Validate(); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("unexpected %d ACH files", len(files))
	}
}

func TestCalculateEffectiveEntryDate(t *testing.T) {
	cfg := config.ODFI{
		Cutoffs: config.Cutoffs{
			Timezone: "America/New_York",
			Windows: []string{
				"14:20",
			},
		},
	}
	timeService := stime.NewStaticTimeService()
	loc, _ := time.LoadLocation(cfg.Cutoffs.Timezone)

	now, _ := time.Parse("2006-01-02 15:04", "2021-04-19 10:00") // layout, value
	timeService.Change(now.In(loc))

	effective := calculateEffectiveEntryDate(cfg, timeService, false)
	if v := effective.String(); v != "2021-04-20 10:00:00 +0000 UTC" {
		t.Error(v)
	}

	// same-day ACH settles today
	effective = calculateEffectiveEntryDate(cfg, timeService, true)
	if v := effective.String(); v != "2021-04-19 10:00:00 +0000 UTC" {
		t.Error(v)
	}

	// advance our timeService
	timeService.Add(5 * time.Hour)
	effective = calculateEffectiveEntryDate(cfg, timeService, false)
	if v := effective.String(); v != "2021-04-21 15:00:00 +0000 UTC" {
		t.Error(v)
	}

	// same day transfers are always the next day
	effective = calculateEffectiveEntryDate(cfg, timeService, true)
	if v := effective.String(); v != "2021-04-20 15:00:00 +0000 UTC" {
		t.Error(v)
	}
}
