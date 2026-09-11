package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	alcsoap "github.com/wz2b/webctrl-soap-go"
)

const LAYOUT = "2006-01-02T15:04:05"

type F1JGroup map[alcsoap.TrendSource]alcsoap.F1JTrendRecord

func main() {
	config, err := processArgs()
	if err != nil {
		log.Fatal(err)
	}

	if config.verbose {
		fmt.Printf(
			"# Start time is %s (%d)\n",
			config.start.Local(),
			config.start.Unix(),
		)

		fmt.Printf(
			"# End time is %s (%d)\n",
			config.stop.Local(),
			config.stop.Unix(),
		)
	}

	//
	// Sources is kept in the order specified by the user/config file.
	// This replaces SortGroup's job.
	//
	var sources []alcsoap.TrendSource
	var columnHeadings []string

	//
	// Groups are indexed by Unix nanoseconds rather than time.Time so
	// timezone/location metadata cannot prevent two equal instants from
	// grouping together.
	//
	groups := make(map[int64]F1JGroup)

	//
	// Current holds the most recent record for each source.  We preload
	// it with the record immediately preceding the requested start time.
	// This replaces GroupRepeatLast's state.
	//
	current := make(map[alcsoap.TrendSource]alcsoap.F1JTrendRecord)

	for _, server := range config.ConfigFile.Servers {
		service := alcsoap.NewSoapService(
			server.Url,
			server.Login,
			server.Password,
		)

		for _, location := range server.Locations {
			source := service.F1JTrend.MakeTrendSource(
				server.Url,
				location.Location,
			)

			sources = append(sources, source)

			if location.Name != "" {
				columnHeadings = append(
					columnHeadings,
					location.Name,
				)
			} else {
				columnHeadings = append(
					columnHeadings,
					location.Location,
				)
			}

			//
			// Get the record immediately before the requested start
			// time so repeat-last has an initial value.
			//
			preload, err := service.F1JTrend.GetF1JTrendData(
				location.Location,
				time.Unix(0, 0),
				config.start,
				false,
				1,
			)

			if err != nil {
				log.Printf(
					"preload %s: %v",
					location.Location,
					err,
				)
			} else if len(preload) > 0 {
				current[source] = preload[len(preload)-1]
			}

			//
			// Get all records in the requested interval.
			//
			records, err := getAllF1JRecords(
				&service.F1JTrend,
				location.Location,
				config.start,
				config.stop,
			)

			if err != nil {
				log.Printf(
					"%s: %v",
					location.Location,
					err,
				)
				continue
			}

			for _, record := range records {
				key := record.Timestamp.UnixNano()

				group, ok := groups[key]
				if !ok {
					group = make(F1JGroup)
					groups[key] = group
				}

				//
				// This preserves the old GroupByTime behavior:
				// one record per source per timestamp.
				//
				// If F1J gives us two records for the same source
				// at exactly the same timestamp, keep the one with
				// the higher sequence number.
				//
				existing, exists := group[source]

				if !exists ||
					record.SequenceNumber >= existing.SequenceNumber {

					group[source] = record
				}
			}
		}
	}

	//
	// Sort all distinct timestamps.
	//
	times := make([]int64, 0, len(groups))

	for timestamp := range groups {
		times = append(times, timestamp)
	}

	sort.Slice(
		times,
		func(i, j int) bool {
			return times[i] < times[j]
		},
	)

	//
	// Each trend now gets all six F1J/CoreTrendRecord fields.
	//
	fmt.Printf("%q", "time")

	fields := []string{
		"sequenceNumber",
		"timestamp",
		"valueType",
		"rawValue",
		"statusFlags",
		"filterDate",
	}

	for _, heading := range columnHeadings {
		for _, field := range fields {
			fmt.Printf(
				"\t%q",
				heading+"."+field,
			)
		}
	}

	fmt.Println()

	//
	// Walk chronologically through the groups.
	//
	// Every record that occurs at this timestamp updates "current".
	// We then output current for every source, which gives us the same
	// repeat-last behavior as the old TrendGrouper.
	//
	for _, timestamp := range times {
		group := groups[timestamp]

		for source, record := range group {
			current[source] = record
		}

		groupTime := time.Unix(0, timestamp)

		fmt.Printf(
			"%q",
			groupTime.Local().
				Format("2006-01-02 15:04:05.000"),
		)

		for _, source := range sources {
			record, ok := current[source]

			if !ok {
				//
				// Six empty columns for this source.
				//
				for i := 0; i < len(fields); i++ {
					fmt.Print("\t")
				}

				continue
			}

			filterDate := ""

			if record.FilterDate != nil {
				filterDate = record.FilterDate.Local().
					Format(time.RFC3339Nano)
			}

			fmt.Printf(
				"\t%d\t%q\t%s\t%q\t%d\t%q",
				record.SequenceNumber,
				record.Timestamp.Local().
					Format(time.RFC3339Nano),
				record.ValueType,
				record.RawValue,
				record.StatusFlags,
				filterDate,
			)
		}

		fmt.Println()
	}
}

// getAllF1JRecords pages through the F1J service.
//
// We're keeping this local to trend_export_f1j for now so none of the
// old TrendService API has to change.
func getAllF1JRecords(
	service *alcsoap.F1JTrendService,
	location string,
	start time.Time,
	stop time.Time,
) ([]alcsoap.F1JTrendRecord, error) {

	const chunkSize = 2000

	var records []alcsoap.F1JTrendRecord

	cursor := start

	for {
		chunk, err := service.GetF1JTrendData(
			location,
			cursor,
			stop,
			true,
			chunkSize,
		)

		if err != nil {
			return nil, err
		}

		if len(chunk) == 0 {
			break
		}

		records = append(records, chunk...)

		if len(chunk) < chunkSize {
			break
		}

		lastTime := chunk[len(chunk)-1].Timestamp

		//
		// F1J timestamps have millisecond resolution.
		//
		next := lastTime.Add(time.Millisecond)

		if !next.After(cursor) {
			return nil, fmt.Errorf(
				"F1J pagination did not advance at %s",
				lastTime,
			)
		}

		if !next.Before(stop) {
			break
		}

		cursor = next
	}

	return records, nil
}
