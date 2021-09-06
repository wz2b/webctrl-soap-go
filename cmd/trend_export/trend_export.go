package main

import (
	"fmt"
	"log"
	"os"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

const LAYOUT = "2006-01-02T15:04:05"

func main() {
	config, err := processArgs()
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
	var sources []*alcsoap.TrendSource

	//
	// Set up the server streams (one per server) while at the same time
	// constructing a flat list of trend locations
	//
	serverStreams := make([]<-chan *alcsoap.TrendEvent, len(config.ConfigFile.Servers))
	for i, server := range config.ConfigFile.Servers {
		service := alcsoap.NewSoapService(server.Url, server.Login, server.Password)

		var locations = make([]string, len(server.Locations))
		for k, location := range server.Locations {
			locations[k] = location.Location
		}

		serverStreams[i] = service.Trend.GetMutlipleTrends(config.start, config.stop, locations)

		for _, location := range server.Locations {
			sources = append(sources, &alcsoap.TrendSource{&service.Trend, location.Location})
		}
	}

	//
	// Output the headers
	//
	fmt.Print("\"time\"\t\"")
	for _, source := range sources {
		fmt.Print(", \"")
		if len(source.Location) > 0 {
			fmt.Print(source.Location)
		} else {
			fmt.Print(source.Location)
		}
	}
	fmt.Println("\"")

	//
	// Run the trends
	//

	merged := alcsoap.MergeTrends(serverStreams)

	grouped := alcsoap.GroupByTime(merged)

	for group := range grouped {
		if group != nil && group.Trends != nil && len(group.Trends) > 0 {

			sorted := alcsoap.SortGroup(group.Trends, sources)
			fmt.Fprintf(os.Stderr, "Sorted list has %d entries\n", len(sorted))

			fmt.Printf("\"%s\"\t", group.Time.Format("2006-01-02 15:04:05"))

			for _, event := range sorted {
				if event != nil {
					fmt.Printf("\t%f", event.Data.Value)
				} else {
					fmt.Printf("\t-")
				}

			}

			fmt.Println()

		} else {
			fmt.Printf("# empty group\n")
		}
	}
}
