package webctrl_soap_go

import (
	"time"
)

type TrendGroupEvent struct {
	Time   time.Time
	Trends []TrendEvent
}

func GroupByTime(events <-chan *TrendEvent) <-chan *TrendGroupEvent {
	var output = make(chan *TrendGroupEvent)

	go func() {
		var group []TrendEvent = nil

		for event := range events {
			//fmt.Println("Input event: ", event.Source.Location, event.Data.Time)
			if event == nil {
				//fmt.Println("Grouper ran out of things to group")
				break
			} else if event.Err != nil {
				//
				// Emit an error with the current time as the timestamp
				//
				newEvent := &TrendGroupEvent{Time: time.Now(), Trends: []TrendEvent{*event}}
				output <- newEvent
			} else if group == nil {
				// First item in list
				//fmt.Println("New group")
				group = []TrendEvent{*event}
			} else if group[0].Data.Time.Equal(event.Data.Time) {
				//fmt.Println("add to group")
				group = append(group, *event)
			} else {
				//fmt.Printf("[%d]", len(group))
				output <- &TrendGroupEvent{Time: group[0].Data.Time, Trends: group}
				group = []TrendEvent{*event}
			}
		}

		//
		// Output the last group
		//
		if group != nil {
			output <- &TrendGroupEvent{Time: group[0].Data.Time, Trends: group}
		}

		//
		// Output an empty record to signal we're done
		//
		close(output)
	}()

	return output
}
