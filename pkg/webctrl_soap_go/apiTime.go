package webctrl_soap_go

import (
	"time"
)

const timePattern = "01/02/2006 03:04:05 PM"

func ParseAlcTime(alcTimeString string) (time.Time, error) {
	tm, err := time.ParseInLocation(timePattern, alcTimeString, time.Local)

	return tm, err
}

func TimeToAlcFormat(instant time.Time) string {
	return instant.Local().Format(timePattern)
}
