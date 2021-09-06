package main

import (
	"time"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

type TrendPointGroup struct {
	Time   *time.Time
	Events map[string]TrendEvent
}

func GroupByTime(events <-chan *alcsoap.TrendEvent, recordsOut chan<- *TrendPointGroup) {
	var group *TrendPointGroup = nil

	for record := <-events; record != nil; record = <-events {
		if group == nil {
			group = &TrendPointGroup{Time: record.Data.Time}
			group.Events = map[string]float32{record.Location: record.Value}
		} else if group.Time.Equal(*record.Time) {
			group.Events[record.Location] = record.Value
		} else {
			recordsOut <- group
			group = &TrendPointGroup{Time: record.Time}
			group.Events = map[string]float32{record.Location: record.Value}
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
