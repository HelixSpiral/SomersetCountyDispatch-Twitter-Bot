package main

import (
	"fmt"

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

	return message
}
