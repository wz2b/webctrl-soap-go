package alcsoap

type TrendMerger struct {
	list *SortedTrendList
}

func CreateTrendMerger() *TrendMerger {
	return &TrendMerger{}
}

func (this *TrendMerger) Merge(channels []<-chan TrendEvent) <-chan TrendEvent {
	var list *SortedTrendList = NewSortedTrendList()

	var output = make(chan TrendEvent)

	go func() {
		// Get the first event from each channel
		for _, channel := range channels {
			// Don't have to check for 'more' here as we are only getting a single point
			event := <-channel
			if event.Err != nil {
				output <- TrendEvent{Source: event.Source, Err: event.Err}
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
			//fmt.Printf("Emitting %s %s\n", event.Data.Time, event.Source.Location)
			output <- event

			// Fetch the next sample off of the same channel this data point came from
			nextEvent := <-item.dataChannel

			// If the result of the last fetch was nil that means there is no more.
			// Otherwise, take this new sample and insert it into the list (merge sorted
			// to the correct place)
			if nextEvent.Err != nil {
				output <- TrendEvent{Source: nextEvent.Source, Err: nextEvent.Err}
			} else if nextEvent.Data.ValueString != "" {
				//fmt.Printf("Inserting %s %s\n", nextEvent.Data.Time, nextEvent.Source.Location)
				list.insert(nextEvent, item.dataChannel)
			}
		}
		close(output)
	}()

	return output
}
