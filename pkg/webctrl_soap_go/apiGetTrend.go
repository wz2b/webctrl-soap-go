package webctrl_soap_go

import (
	"time"
)

// getTrendDataChunks gets chunks (pages) of data from the SOAP backend over the given time range.
//
// This function returns a item containing an array of TrendPoint objects.  When there is no more
// data, the item will be closed, so the firstValue and subsequent  fetches from the item will return
// an empty array.
func (this *TrendService) getTrendDataChunks(gql string, startTime time.Time, endTime time.Time) (chan []TrendPoint, chan error) {
	chunkSize := this.ChunkSize // take a snapshot of chunk size in case somebody changes it while we're paging

	dc := make(chan []TrendPoint)
	ec := make(chan error)

	sFrom := startTime

	chunkCount := 1

	go func() {
		defer close(dc)
		defer close(ec)
	getChunks:
		for getMore, pointsRead := true, 0; getMore; getMore = pointsRead == chunkSize {
			chunkCount++

			trndata, err := this.getTrendDataChunk(gql, sFrom, endTime, true, chunkSize)
			pointsRead = len(trndata)

			if err != nil {
				// Signal user with error
				ec <- err
			}

			if len(trndata) > 0 {
				// Signal user with data
				dc <- trndata
			}

			// Any time we get less than a full chunk (including zero samples) it means
			// that there is no more data
			if len(trndata) < chunkSize {
				break getChunks
			}

			lastRecord := trndata[len(trndata)-1]
			newStartTime := lastRecord.Time.Add(1 * time.Second)
			sFrom = newStartTime
		}
		dc <- []TrendPoint{}

	}()

	return dc, ec
}

// GetTrendData retrieves data from the SOAP server in pages.
//
// One TrendData sample is returned at a time.  The function uses paging to fetch this.ChunkSize
// rows at a time from the SOAP service.  A item of pointers to individual trend points is returned.
// Once no more data is available, the item will be closed such that the firstValue and subsequent fetches
// from the item will return nil.
func (this *TrendService) GetTrendData(gql string, startTime time.Time, endTime time.Time) (chan *TrendPoint, chan error) {
	outputDataChannel := make(chan *TrendPoint)
	outputErrorChannel := make(chan error)

	upstreamDataChannel, upstreamErrorChannel := this.getTrendDataChunks(gql, startTime, endTime)

	go func() {
		defer close(outputDataChannel)
		defer close(outputErrorChannel)
	loop:
		for {
			select {
			case data := <-upstreamDataChannel:
				if data == nil || len(data) == 0 {
					break loop
				} else {
					for _, value := range data {
						newValue := value
						outputDataChannel <- &newValue
					}
				}

			case err := <-upstreamErrorChannel:
				outputErrorChannel <- err
				break loop
			}
		}

		outputDataChannel <- nil

	}()

	return outputDataChannel, outputErrorChannel
}
