package alcsoap

import (
	"time"
)

// GetMultipleTrends returns a merge-sorted stream of trend point events.
func (this *TrendService) GetMultipleTrends(startTime time.Time, stopTime time.Time, locations []string) <-chan TrendEvent {

	merger := CreateTrendMerger()
	var channels = make([]<-chan TrendEvent, len(locations))
	for i, location := range locations {
		channels[i] = this.GetTrendData(location, startTime, stopTime)
	}

	return merger.Merge(channels)
}
