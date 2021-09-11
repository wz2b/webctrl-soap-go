package webctrl_soap_go

import (
	"time"
)

type TrendSource struct {
	Server   string
	Location string
}

type TrendEvent struct {
	Source TrendSource
	Data   TrendPoint
	Err    error
}

// GetTrendData gets trend data, paging if necessary, and returns a channel that emits a stream
// of trend points.
//
// This function returns a item containing an array of TrendPoint objects.  When there is no more
// data, the item will be closed, so the firstValue and subsequent  fetches from the item will return
// an empty array.
func (this *TrendService) GetTrendData(gql string, startTime time.Time, endTime time.Time) <-chan TrendEvent {
	chunkSize := this.chunkSize // take a snapshot of chunk size in case somebody changes it while we're paging
	sFrom := startTime

	source := TrendSource{Server: this.parent.Name, Location: gql}

	events := make(chan TrendEvent)

	go func() {
	getChunks:
		for getMore, pointsRead := true, 0; getMore; getMore = pointsRead == chunkSize {
			// This is a synchronous SOAP call
			trnData, err := this.getTrendDataChunk(gql, sFrom, endTime, true, chunkSize)
			pointsRead = len(trnData)

			if err != nil {
				events <- TrendEvent{Err: err, Source: source}
			}

			if len(trnData) > 0 {
				// Signal user with data
				for _, point := range trnData {
					events <- TrendEvent{Data: TrendPoint{
						Time:        point.Time,
						TimeString:  point.TimeString,
						ValueString: point.ValueString,
						Value:       point.Value,
					}, Err: nil, Source: source}
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
		//fmt.Println("Closing channel for emitting", gql)
		close(events)
	}()
	return events
}

func (this *TrendService) MakeTrendSource(server string, location string) TrendSource {
	return TrendSource{Server: server, Location: location}
}
