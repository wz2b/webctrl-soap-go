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
	sFrom := startTime

	chunkCount := 1

	outData := make(chan []TrendPoint)
	outErrors := make(chan error)

	go func() {
		defer close(outData)
		defer close(outErrors)
	getChunks:
		for getMore, pointsRead := true, 0; getMore; getMore = pointsRead == chunkSize {
			chunkCount++

			// This is a synchronous SOAP call
			trnData, err := this.getTrendDataChunk(gql, sFrom, endTime, true, chunkSize)
			pointsRead = len(trnData)

			if err != nil {
				// Signal user with error
				outErrors <- err
			}

			if len(trnData) > 0 {
				// Signal user with data
				outData <- trnData
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
		outData <- []TrendPoint{}
	}()

	return outData, outErrors
}

// GetTrendData retrieves data from the SOAP server in pages.
//
// One TrendData sample is returned at a time.  The function uses paging to fetch this.ChunkSize
// rows at a time from the SOAP service.  A item of pointers to individual trend points is returned.
// Once no more data is available, the item will be closed such that the firstValue and subsequent fetches
// from the item will return nil.
func (this *TrendService) GetTrendData(gql string, startTime time.Time, endTime time.Time) (chan *TrendPoint, chan error) {
	dataCh := make(chan *TrendPoint)
	errorCh := make(chan error)
	chunkDataChannel, chunkErrorChannel := this.getTrendDataChunks(gql, startTime, endTime)

	go func() {
		defer close(dataCh)
		defer close(errorCh)
	loop:
		for {
			select {
			case data := <-chunkDataChannel:
				if data == nil || len(data) == 0 {
					break loop
				} else {
					for _, value := range data {
						newValue := value
						dataCh <- &newValue
					}
				}

			case err := <-chunkErrorChannel:
				errorCh <- err
				break loop
			}
		}

		dataCh <- nil
	}()

	return dataCh, errorCh
}
