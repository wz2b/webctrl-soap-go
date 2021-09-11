package main

import (
	"fmt"
	"log"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

const LAYOUT = "2006-01-02T15:04:05"

func main() {
	config, err := processArgs()
	grouper := alcsoap.CreateTrendGrouper()

	if err != nil {
		log.Fatal(err)
	}

	if config.verbose {
		fmt.Printf("# Start time is %s (%d)\n", config.start.Local(), config.start.Unix())
		fmt.Printf("# End time is %s (%d)\n", config.stop.Local(), config.stop.Unix())
	}

	//
	// Make a flattened list of servers
	//
	var sources []alcsoap.TrendSource

	//
	// Set up the server streams (one per server) while at the same time
	// constructing a flat list of trend locations
	//
	serverStreams := make([]<-chan alcsoap.TrendEvent, len(config.ConfigFile.Servers))
	//fmt.Printf("There are %d servers\n", len(config.ConfigFile.Servers))

	var columnHeadings = make([]string, 0)

	for i, server := range config.ConfigFile.Servers {
		service := alcsoap.NewSoapService(server.Url, server.Login, server.Password)

		var locations = make([]string, len(server.Locations))
		for k, location := range server.Locations {
			locations[k] = location.Location
		}

		serverStreams[i] = service.Trend.GetMultipleTrends(config.start, config.stop, locations)

		for _, location := range server.Locations {
			sourceLocation := service.Trend.MakeTrendSource(server.Url, location.Location)
			sources = append(sources, sourceLocation)
			if len(location.Name) > 0 {
				columnHeadings = append(columnHeadings, location.Name)
			} else {
				columnHeadings = append(columnHeadings, location.Location)
			}

			grouper.Preload(service.Trend, config.start, sourceLocation)

		}

	}

	//
	// Output the headers
	//
	fmt.Print("\"time")
	for _, heading := range columnHeadings {
		fmt.Printf("\"\t \"%s", heading)
	}
	fmt.Println("\"")

	//
	// TODO: when implementing the interpolation module, build in a way to always find the first point

	// TODO: hide password from TrendSource when outputting.  Upsetting how it shows up.

	//
	// Run the trends
	//
	merger := alcsoap.CreateTrendMerger()

	merged := merger.Merge(serverStreams)

	grouped := grouper.GroupByTime(merged)
	repeated := grouper.GroupRepeatLast(grouped)

	for group := range repeated {
		sorted := alcsoap.SortGroup(group, sources)
		fmt.Printf("\"%s\"\t", group.Time.Format("2006-01-02 15:04:05"))

		for _, event := range sorted.Trends {
			if event.Err != nil {
				fmt.Printf("\tErr")
			} else if event.Data.IsValid() == false {
				fmt.Printf("\t-")
			} else {
				fmt.Printf("\t%f", event.Data.Value)
			}

		}

		fmt.Println()

	}
}
