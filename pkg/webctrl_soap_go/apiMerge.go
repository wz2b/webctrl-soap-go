package webctrl_soap_go

import (
	"time"
)

func (this *TrendService) MergeSortTrends(startTime time.Time, stopTime time.Time, locations []string) <-chan *TrendEvent {
	var list = &sortedList{head: nil}
	output := make(chan *TrendEvent)

	go func() {
		for _, location := range locations {
			dataCh := this.GetTrendData(location, startTime, stopTime)

			firstPoint := <-dataCh

			if firstPoint != nil {
				//fmt.Printf("Insert %s\n", firstPoint.Time.Local())
				list.insert(firstPoint, dataCh)
			}
		}

		for item := list.popFirst(); item != nil; item = list.popFirst() {
			event := item.event

			output <- event

			// Fetch the next sample off of whatever one we pulled this data from
			nextValue := <-item.dataChannel

			if nextValue != nil {
				if nextValue.err != nil {
					output <- &TrendEvent{
						Source: event.Source,
						Data:   nil,
						err:    nextValue.err,
					}
				} else {
					// If the result of the last fetch was nil that means there is no more.
					// Otherwise, take this new sample and insert it into the list (merge sorted
					// to the corret place)
					list.insert(nextValue, item.dataChannel)
				}
			}
		}

		close(output)
	}()
	return output
}
