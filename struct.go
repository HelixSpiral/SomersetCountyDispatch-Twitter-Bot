package main

import (
	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
)

type Cache struct {
	Day           int
	LastProcessed string                                         // The last CallNum we have processed
	LogMap        map[string][]somersetcountywrapper.DispatchLog // Our cache
}

type MqttMessage struct {
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	TwitterAccessToken    string
	TwitterAccessSecret   string

	Message string
	Image   string
}
