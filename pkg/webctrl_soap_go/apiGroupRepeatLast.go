package webctrl_soap_go

import (
	"time"
)

func (this *TrendGrouper) Preload(svc TrendService, start time.Time, location TrendSource) {

	beginningOfTime := time.Unix(0, 0)

	// func (this *TrendService) getTrendDataChunk(gql string, startTime time.Time, endTime time.Time, fromStart bool, maxRecords int) ([]TrendPoint, error) {
	points, err := svc.getTrendDataChunk(location.Location, beginningOfTime, start, false, 1)
	if err == nil {
		for _, pt := range points {
			this.preload[location] = TrendEvent{Source: location, Data: pt, Err: nil}
		}
	}
}

func (this *TrendGrouper) GroupRepeatLast(input <-chan TrendGroupEvent) <-chan TrendGroupEvent {

	output := make(chan TrendGroupEvent)

	go func() {

		last := make(map[TrendSource]TrendEvent)

		for k, v := range this.preload {
			last[k] = v
		}

		for ingrp := range input {
			if ingrp.IsEmpty() {
				break
			}

			outgrp := make(map[TrendSource]TrendEvent)

			// All the new fields will update last
			for key, entry := range ingrp.Trends {
				if entry.Data.ValueString != "" {
					last[key] = entry
				}
			}

			// Now, everything in 'last' is what we want to output
			for key, entry := range last {
				outgrp[key] = entry
			}

			output <- TrendGroupEvent{Time: ingrp.Time, Trends: outgrp}
		}

		close(output)
	}()

	return output
}
