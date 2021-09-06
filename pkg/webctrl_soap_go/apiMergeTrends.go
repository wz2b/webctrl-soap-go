package webctrl_soap_go

import "errors"

func MergeTrends(channels []<-chan *TrendEvent) <-chan *TrendEvent {
	var list = &sortedList{head: nil}

	var output = make(chan *TrendEvent)

	go func() {
		// Get the first event from each channel
		for _, channel := range channels {
			// Don't have to check for 'more' here as we are only getting a single point
			event := <-channel

			if event == nil {
				output <- &TrendEvent{
					Source: event.Source, Data: nil, Err: errors.New("trend source is empty")}
			} else if event.Err != nil {
				output <- &TrendEvent{Source: event.Source, Data: nil, Err: event.Err}
			} else {
				list.insert(event, channel)
			}
		}

		//
		// Pop the sorted list until it is empty.  Every time an item is removed,
		// fetch another entry from that channel and re-insert it in order
		//
		for item := list.popFirst(); item != nil; item = list.popFirst() {
			event := item.event
			output <- event

			// Fetch the next sample off of the same channel this data point came from
			nextEvent := <-item.dataChannel

			// If the result of the last fetch was nil that means there is no more.
			// Otherwise, take this new sample and insert it into the list (merge sorted
			// to the corret place)
			if nextEvent != nil {
				if event.Err != nil {
					output <- &TrendEvent{Source: event.Source, Data: nil, Err: event.Err}
				} else {
					list.insert(nextEvent, item.dataChannel)
				}
			}
		}
		close(output)
	}()

	return output
}
