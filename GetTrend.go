package webctrl_soap_go

import (
	"time"
)

func (this *TrendService) GetTrendData(gql string, startTime time.Time, endTime time.Time, chunkSize int) (chan []TrendPoint, chan error) {
	dc := make(chan []TrendPoint)
	ec := make(chan error)

	sFrom := startTime

	chunkCount := 1

	go func() {
		// Closing the error channel will tell listeners we're done
		defer close(ec)

		getChunks: for getMore, pointsRead := true, 0; getMore; getMore = pointsRead == chunkSize {
			chunkCount++

			trndata, err := this.GetTrendChunk(gql, sFrom, endTime, true, chunkSize)
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
	}()

	return dc, ec
}
