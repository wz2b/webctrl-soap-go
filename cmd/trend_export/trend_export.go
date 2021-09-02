package main

import (
	"fmt"
	"log"
	"strings"
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

	alc := alcsoap.NewSoapService(config.server, config.user, config.password)

	alc.Trend.ChunkSize = 1000

	//for _, arg := range config.args {
	//	data, err := alc.Trend.GetTrendData(arg, config.start, config.stop)
	//loop:
	//	for {
	//		select {
	//		case d := <-data:
	//			if d != nil {
	//				fmt.Printf("%s\t%f\n", d.Time.Local(), d.Value)
	//			}
	//
	//		case e := <-err:
	//			if e != nil {
	//				log.Print(e)
	//			}
	//			break loop
	//		}
	//	}
	//}

	fields := config.args
	var records = make(chan *alcsoap.TrendPoint)
	go alc.Trend.MergeTendData(fields, config.start, config.stop, records)

	var grouped = make(chan *alcsoap.TrendPointGroup)
	go alcsoap.GroupByTime(records, grouped)

	fmt.Printf("\"time\"\t\"" + strings.Join(fields, "\"\t\"") + "\"\n")

	for group := <-grouped; group != nil; group = <-grouped {
		fmt.Printf("\"%s\"\t", group.Time.Local().Format("2006-01-02 15:04:05"))

		for _, field := range fields {
			value, ok := group.Points[field]
			if ok {
				fmt.Printf("%f\t", value)
			} else {
				fmt.Printf("\t")
			}
		}

		fmt.Println()
	}
}
