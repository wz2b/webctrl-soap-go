package webctrl_soap_go

import (
	"time"
)

type TrendGroupEvent struct {
	Time   time.Time
	Trends map[TrendSource]TrendEvent
}

func (this *TrendGroupEvent) IsEmpty() bool {
	return len(this.Trends) == 0
}

type OrderedTrendGroupEvent struct {
	Time   time.Time
	Trends []TrendEvent
}

func (this *OrderedTrendGroupEvent) IsEmpty() bool {
	return len(this.Trends) == 0
}

type TrendGrouper struct {
	preload map[TrendSource]TrendEvent
}

func CreateTrendGrouper() *TrendGrouper {
	return &TrendGrouper{
		preload: make(map[TrendSource]TrendEvent),
	}
}

func (this *TrendGrouper) GroupByTime(events <-chan TrendEvent) <-chan TrendGroupEvent {
	var output = make(chan TrendGroupEvent)

	currentGroup := make(map[TrendSource]TrendEvent)

	go func() {
		var groupTime time.Time

		for event := range events {
			if event.Err != nil {
				//
				// Emit an error with the current time as the timestamp
				//
				newEvent := TrendGroupEvent{Time: time.Now(), Trends: currentGroup}
				output <- newEvent
			} else if event.Data.IsValid() == false {
				break
			} else if len(currentGroup) == 0 {
				// First item in list
				currentGroup[event.Source] = event
				groupTime = event.Data.Time
			} else if groupTime.Equal(event.Data.Time) {
				currentGroup[event.Source] = event
			} else {
				//
				// End this group
				//
				output <- TrendGroupEvent{Time: groupTime, Trends: currentGroup}
				currentGroup = make(map[TrendSource]TrendEvent)
				currentGroup[event.Source] = event
				groupTime = event.Data.Time
			}
		}

		//
		// Output the last group
		//
		if len(currentGroup) > 0 {
			output <- TrendGroupEvent{Time: groupTime, Trends: currentGroup}
		}

		//
		// Output an empty record to signal we're done
		//
		close(output)
	}()

	return output
}
