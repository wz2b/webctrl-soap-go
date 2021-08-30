package main

import (
	"flag"
	"fmt"
	webctrl_soap_go "github.com/wz2b/webctrl-soap-go"
	"log"
	"time"
)

const LAYOUT = "2006-01-02T15:04:05"

func main() {
	user := flag.String("user", "", "WebCTRL username (must have SOAP privileges)")
	password := flag.String("password", "", "WebCTRL password")
	url := flag.String("url", "", "URL to WebCTRL server")
	startStr := flag.String("start", "", "Start time (inclusive)")
	stopStr := flag.String("stop", "", "End time (exclusive)")
	verbose := flag.Bool("v", false, "Display extra information")

	flag.Parse()

	if *user == "" || *password == "" || *url == "" {
		log.Fatal("user, password, and url are all required arguments")
	}

	var start time.Time
	var stop time.Time
	var err error

	if *startStr == "" {
		start = time.Now().Add(-24 * time.Hour)
	} else {
		start, err = time.ParseInLocation(LAYOUT, *startStr, time.Local)
		if err != nil {
			log.Fatal("Could not parse start time")
		}
	}

	if *stopStr == "" {
		// If the user doesn't specify a time, use 1 second from now - that will get
		// everything up to and including now (if such a sample is available).
		stop = time.Now().Add(1 * time.Second)
	} else {
		stop, err = time.ParseInLocation(LAYOUT, *stopStr, time.Local)
		if err != nil {
			log.Fatal("Could not parse start time")
		}
	}

	args := flag.Args()

	if(*verbose) {
		fmt.Printf("# Start time is %s (%d)\n", start.Local(), start.Unix())
		fmt.Printf("# End time is %s (%d)\n", stop.Local(), stop.Unix())
	}

	alc := webctrl_soap_go.NewSoapService(*url, *user, *password)

	for _, arg := range args {
		data, err := alc.Trend.GetTrendData(arg, start, stop, 100)

		chunks := 0
		loop: for
		{
			select {
			case d := <-data:
				chunks++
				if(*verbose) {
					fmt.Printf("# chunk %d\n", chunks)
				}
				for _, pt := range(d) {
					fmt.Printf("%s\t%f\n", pt.Time.Local(), pt.Value)
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
