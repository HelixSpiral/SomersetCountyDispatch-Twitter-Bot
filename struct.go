package main

import (
	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
)

type Cache struct {
	Day                  int
	LastProcessed        string // The last CallNum we have processed
	XAppLimit24HourReset int64
	XAppRateLimited      bool
	XRateLimitReset      int64
	XRateLimited         bool
	LogMap               map[string][]somersetcountywrapper.DispatchLog // Our cache
}
