package main

import (
	"fmt"
	"github.com/araddon/dateparse"
	"testing"
	"time"
)

func TestParsing(t *testing.T) {
	kievTime, _ := time.LoadLocation("Europe/Kiev")
	utcTime, _ := time.LoadLocation("UTC")

	dt, _ := dateparse.ParseAny("2019-03-10 22:01:00 UTC")
	dtStr := fmt.Sprintf("%s", dt)
	if "2019-03-10 22:01:00 +0000 UTC" != dtStr {
		t.Errorf("Datetime parsed incorrectly: got %s", dtStr)
	}

	dt, _ = dateparse.ParseAny("2019-03-10 22:01:00")
	dtStr = fmt.Sprintf("%s", dt)
	if "2019-03-10 22:01:00 +0000 UTC" != dtStr {
		t.Errorf("Datetime parsed incorrectly: got %s", dtStr)
	}
	dtStr = fmt.Sprintf("%s", dt.In(kievTime))
	if "2019-03-11 00:01:00 +0200 EET" != dtStr {
		t.Errorf("Datetime parsed incorrectly: got %s", dtStr)
	}

	dt, _ = dateparse.ParseAny("2019-03-11 00:01:00 +0200 EET")
	dtStr = fmt.Sprintf("%s", dt)
	if "2019-03-11 00:01:00 +0200 EET" != dtStr {
		t.Errorf("Datetime parsed incorrectly: got %s", dtStr)
	}
	dtStr = fmt.Sprintf("%s", dt.In(utcTime))
	if "2019-03-10 22:01:00 +0000 UTC" != dtStr {
		t.Errorf("Datetime parsed incorrectly: got %s", dtStr)
	}
}
