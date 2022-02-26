// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package upload

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/moov-io/paygate/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig__OutboundFilenameTemplate(t *testing.T) {
	var cfg *config.ODFI
	if tmpl := cfg.FilenameTemplate(); tmpl != config.DefaultFilenameTemplate {
		t.Errorf("expected default template: %v", tmpl)
	}

	cfg = &config.ODFI{
		OutboundFilenameTemplate: `{{ date "20060102" }}`,
	}
	if tmpl := cfg.FilenameTemplate(); tmpl == config.DefaultFilenameTemplate {
		t.Errorf("expected custom template: %v", tmpl)
	}
}

func TestFilenameTemplate(t *testing.T) {
	// default
	filename, err := RenderACHFilename(config.DefaultFilenameTemplate, FilenameData{
		RoutingNumber: "987654320",
		GPG:           true,
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	yymmdd := now.Format("20060102")
	hhmm := now.Format("1504")
	expected := fmt.Sprintf("%s-%s-987654320.ach.gpg", yymmdd, hhmm)
	if filename != expected {
		t.Errorf("filename=%s", filename)
	}

	// example from original issue
	linden := `{{ date "20060102" }}.ach`
	filename, err = RenderACHFilename(linden, FilenameData{
		// not included in template
		GPG:           true,
		RoutingNumber: "987654320",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected = fmt.Sprintf("%s.ach", time.Now().Format("20060102"))
	if filename != expected {
		t.Errorf("filename=%s", filename)
	}
}

func TestFilenameTemplate__functions(t *testing.T) {
	cases := []struct {
		tmpl, expected string
		data           FilenameData
	}{
		{
			tmpl:     "static-template",
			expected: "static-template",
		},
		{
			tmpl:     `{{ env "PATH" }}`,
			expected: os.Getenv("PATH"),
		},
		{
			tmpl:     `{{ date "2006-01-02" }}`,
			expected: time.Now().Format("2006-01-02"),
		},
	}
	for i := range cases {
		res, err := RenderACHFilename(cases[i].tmpl, cases[i].data)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
		}
		if cases[i].expected != res {
			t.Errorf("#%d: %s", i, res)
		}
	}
}

func TestFilenameTemplate__RoundSequenceNumber(t *testing.T) {
	if n := RoundSequenceNumber(0); n != "0" {
		t.Errorf("got %s", n)
	}
	if n := RoundSequenceNumber(10); n != "A" {
		t.Errorf("got %s", n)
	}
}

func TestFilenameTemplate__ACHFilenameSeq(t *testing.T) {
	n := ACHFilenameSeq("")
	assert.Equal(t, n, 0)

	n = ACHFilenameSeq("20210102-C.ach")
	assert.Equal(t, n, 12)

	n = ACHFilenameSeq("20060102-0830-987654320-1.ach")
	assert.Equal(t, n, 1)

	n = ACHFilenameSeq("20060102-987654320-1.ach")
	assert.Equal(t, n, 1)

	n = ACHFilenameSeq("20060102-987654320-2.ach.gpg")
	assert.Equal(t, n, 2)

	n = ACHFilenameSeq("my-20060102-987654320-3.ach")
	assert.Equal(t, n, 3)

	n = ACHFilenameSeq("20060102-B-987654320.ach")
	assert.Equal(t, n, 11)
}
