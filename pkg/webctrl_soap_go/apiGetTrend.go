package webctrl_soap_go

import (
	"time"
)

type TrendEvent struct {
	Source *TrendService
	Data   *TrendPoint
	err    error
}

// GetTrendData gets trend data, paging if necessary, and returns a channel that emits a stream
// of trend points.
//
// This function returns a item containing an array of TrendPoint objects.  When there is no more
// data, the item will be closed, so the firstValue and subsequent  fetches from the item will return
// an empty array.
func (this *TrendService) GetTrendData(gql string, startTime time.Time, endTime time.Time) <-chan *TrendEvent {
	chunkSize := this.ChunkSize // take a snapshot of chunk size in case somebody changes it while we're paging
	sFrom := startTime

	chunkCount := 1

	events := make(chan *TrendEvent)

	go func() {
		defer close(events)
	getChunks:
		for getMore, pointsRead := true, 0; getMore; getMore = pointsRead == chunkSize {
			chunkCount++

			// This is a synchronous SOAP call
			trnData, err := this.getTrendDataChunk(gql, sFrom, endTime, true, chunkSize)
			pointsRead = len(trnData)

			if err != nil {
				// Signal user with error
				events <- &TrendEvent{Data: nil, err: err, Source: this}
			}

			if len(trnData) > 0 {
				// Signal user with data
				for _, point := range trnData {
					events <- &TrendEvent{Data: &point, err: nil, Source: this}
				}
			}

			// Any time we get less than a full chunk (including zero samples) it means
			// that there is no more data
			if len(trnData) < chunkSize {
				break getChunks
			}

			lastRecord := trnData[len(trnData)-1]
			newStartTime := lastRecord.Time.Add(1 * time.Second)
			sFrom = newStartTime
		}
		events <- nil
	}()

	return events
}
