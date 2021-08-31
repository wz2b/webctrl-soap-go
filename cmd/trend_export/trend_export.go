package main

import (
	"fmt"
	"log"
	"webctrl-soap-go/pkg/webctrl_soap_go"
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

	alc := webctrl_soap_go.NewSoapService(config.server, config.user, config.password)

	alc.Trend.ChunkSize = 100

	for _, arg := range config.args {
		data, err := alc.Trend.GetTrendData(arg, config.start, config.stop)
	loop:
		for {
			select {
			case d := <-data:
				if d != nil {
					fmt.Printf("%s\t%f\n", d.Time.Local(), d.Value)
				}

			case e := <-err:
				if e != nil {
					log.Print(e)
				}
				break loop
			}
		}
	}
}
