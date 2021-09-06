package webctrl_soap_go

import "errors"

func MergeSortMultipleTrends(channels []<-chan *TrendEvent) chan<- *TrendEvent {
	var list = &sortedList{head: nil}

	var output = make(chan<- *TrendEvent)

	go func() {
		// Get the first event from each channel
		for _, channel := range channels {
			event := <-channel

			if event == nil {
				output <- &TrendEvent{
					Source: event.Source, Data: nil, err: errors.New("trend source is empty")}
			} else if event.err != nil {
				output <- &TrendEvent{Source: event.Source, Data: nil, err: event.err}
			} else {
				list.insert(event, channel)
			}
		}

		//
		// Run all trends as a merge sort
		//
		for item := list.popFirst(); item != nil; item = list.popFirst() {
			event := item.event
			output <- event

			// Fetch the next sample off of whatever one we pulled this data from
			nextEvent := <-item.dataChannel

			// If the result of the last fetch was nil that means there is no more.
			// Otherwise, take this new sample and insert it into the list (merge sorted
			// to the corret place)
			if nextEvent != nil {
				if nextEvent == nil {
					output <- &TrendEvent{
						Source: nextEvent.Source, Data: nil, err: errors.New("trend source is empty")}
				} else if event.err != nil {
					output <- &TrendEvent{Source: event.Source, Data: nil, err: event.err}
				} else {
					list.insert(nextEvent, item.dataChannel)
				}
			}
		}
		output <- nil

	}()

	return output
}
