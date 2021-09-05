package main

import (
	"fmt"
	"log"
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
	type flat struct {
		server   ServerConfig
		location TrendLocation
	}

	var flattened []flat

	for _, server := range config.ConfigFile.Servers {
		for _, location := range server.Locations {
			flattened = append(flattened, flat{server, location})
		}
	}

	//
	// Output the headers
	//
	fmt.Print("\"time\"\t\"")
	for _, flattened := range flattened {
		fmt.Print(", \"")
		if len(flattened.location.Location) > 0 {
			fmt.Print(flattened.location.Name)
		} else {
			fmt.Print(flattened.location.Location)
		}
	}
	fmt.Println("\"")

	//
	// Run the trends
	//
	records := make(chan *alcsoap.TrendPoint)
	go MergeSortTrendData(config.ConfigFile.Servers, config.start, config.stop, records)

	var grouped = make(chan *TrendPointGroup)
	go GroupByTime(records, grouped)

	for group := <-grouped; group != nil; group = <-grouped {
		fmt.Printf("\"%s\"\t", group.Time.Local().Format("2006-01-02 15:04:05"))

		for _, loc := range flattened {
			value, ok := group.Points[loc.location.Location]
			if ok {
				fmt.Printf("%f\t", value)
			} else {
				fmt.Printf("-\t")
			}
		}

		fmt.Println()
	}
}
