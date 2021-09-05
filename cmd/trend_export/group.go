package main

import (
	"time"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

type TrendPointGroup struct {
	Time   *time.Time
	Points map[string]float32
}

func GroupByTime(recordsIn <-chan *alcsoap.TrendPoint, recordsOut chan<- *TrendPointGroup) {
	var group *TrendPointGroup = nil

	for record := <-recordsIn; record != nil; record = <-recordsIn {
		if group == nil {
			group = &TrendPointGroup{Time: record.Time}
			group.Points = map[string]float32{record.Location: record.Value}
		} else if group.Time.Equal(*record.Time) {
			group.Points[record.Location] = record.Value
		} else {
			recordsOut <- group
			group = &TrendPointGroup{Time: record.Time}
			group.Points = map[string]float32{record.Location: record.Value}
		}
	}

	//
	// Output the last group
	//
	if group != nil {
		recordsOut <- group
	}

	//
	// Output an empty record to signal we're done
	//
	recordsOut <- nil
}
