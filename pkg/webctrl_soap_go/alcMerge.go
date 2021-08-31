package webctrl_soap_go

import (
	"fmt"
	"reflect"
	"time"
)

type channelSpec struct {
	dataChannel  chan *TrendPoint
	errorChannel chan error
	gql          string
}



func (this *TrendService) MergeTendData(gqlPaths []string, startTime time.Time, endTime time.Time) (chan *TrendPoint, chan error) {
	//out := make(chan *TrendPoint)
	channelSpecs := make([]*channelSpec, len(gqlPaths))

	for i, gql := range gqlPaths {
		dataChannel, errorChannel := this.GetTrendData(gql, startTime, endTime)
		channelSpecs[i] = &channelSpec{dataChannel: dataChannel, errorChannel: errorChannel, gql: gql}
	}

	dataChannels := make([]reflect.SelectCase, len(channelSpecs))
	errorChannels := make([]reflect.SelectCase, len(channelSpecs))
	for i, channel := range channelSpecs {
		dataChannels[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel.dataChannel)}
		errorChannels[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel.errorChannel)}
	}

	for {
		//
		// nChosen is the index of the chosen case
		// receiveOk indicates true if a value was returned and false if a channel was closed
		//
		nChosen, receivedValue, recvOk := reflect.Select(dataChannels)


		fmt.Printf("%d: %s %s\n", nChosen, receivedValue, recvOk)




	}
}
