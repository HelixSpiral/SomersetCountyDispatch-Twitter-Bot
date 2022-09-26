package main

import (
	"fmt"
	"strings"

	somersetcountywrapper "github.com/HelixSpiral/SomersetCountyAPIWrapper"
)

func buildMessage(log somersetcountywrapper.DispatchLog) string {
	var message string

	message = fmt.Sprintf("[%s/%s] Reason: %s", log.CallNum, log.CallTime, log.ReasonText)

	if log.Jurisdiction != "" {
		message += fmt.Sprintf(" | Location: [%s]", log.Jurisdiction)

		if log.StreetName != "" {
			message += fmt.Sprintf(" %s", log.StreetName)

			if log.StreetSuf != "" {
				message += fmt.Sprintf(" %s", log.StreetSuf)
			}
		}
	}

	switch log.UnitType {
	case "F":
		if log.UnitDesc != "" {
			message += fmt.Sprintf(" | Fire Unit: %s", log.UnitDesc)
		}
	case "E":
		if log.UnitDesc != "" {
			message += fmt.Sprintf(" | EMS Unit: %s", log.UnitDesc)
		}
	case "P":
		if log.Unit != "" {
			message += fmt.Sprintf(" | Police Unit: %s", log.Unit)
		}
	}

	// Should probably move this out into a config file at some point so we can update it
	// without a code change and rebuild of the app.
	tagWords := []string{
		"Police",
		"EMS",
		"Fire",
		"MEDICAL",
		"EMERGENCY",
		"THEFT",
		"ACCIDENT",
		"Complaint",
		"DOMESTIC",
		"ANIMAL",
		"SUSPICIOUS",
		"Welfare",
		"Animal",
		"Mischief",
		"CITIZEN",
		"Wildlife",
		"SHOPLIFTING",
		"DISTURBANCE",
		"SMOKE",
		"BURGLARY",
		"TRESSPASS",
	}

	for _, y := range tagWords {
		message = strings.Replace(message, fmt.Sprintf(" %s ", y), fmt.Sprintf(" #%s ", y), 1)
	}

	return message
}
