package main

import (
	"fmt"
	"log"
	"time"
	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"
)

type channelSpec struct {
	dataChannel  chan *alcsoap.TrendPoint
	errorChannel chan error
	location     TrendLocation
	firstValue   *alcsoap.TrendPoint
}

type sortedList struct {
	head  *listItem
	count uint
}
type listItem struct {
	item *channelSpec
	next *listItem
}

// insert() inserts an item into a singly linked list based on the timestamp.
func (list *sortedList) insert(newEntry *channelSpec) *sortedList {
	newNode := &listItem{item: newEntry, next: nil}

	if list.head == nil {
		// List is empty - this item is the head
		newNode.next = nil
		list.head = newNode
	} else if newEntry.firstValue.Time.Before(*list.head.item.firstValue.Time) {
		// This item goes before the first item
		newNode.next = list.head
		list.head = newNode
	} else {
		// Find where the new item goes and insert it then link to the current location
		var current = list.head
		for current.next != nil && newEntry.firstValue.Time.After(*current.item.firstValue.Time) {
			current = current.next
		}

		newNode.next = current.next
		current.next = newNode
	}

	list.count = list.count + 1
	return list
}

func (list *sortedList) popFirst() *listItem {
	item := list.head

	if item != nil {
		list.head = item.next
		list.count = list.count - 1
	}

	return item
}

func MergeSortTrendData(servers []ServerConfig, startTime time.Time, endTime time.Time, output chan<- *alcsoap.TrendPoint) {
	var list = &sortedList{head: nil}

	//
	// Set up the SOAP get trend calls and push the first data point
	// from each trend
	//
	for _, server := range servers {
		for _, location := range server.Locations {
			fmt.Printf("Setup for %s %s\n", server.Url, location.Location)
			alc := alcsoap.NewSoapService(server.Url, server.Login, server.Password)
			trnServer := alc.Trend
			trnServer.ChunkSize = 1000
			dataCh, errCh := trnServer.GetTrendData(location.Location, startTime, endTime)
			firstPoint := <-dataCh
			eatError(errCh)

			if firstPoint != nil {
				//fmt.Printf("Insert %s\n", firstPoint.Time.Local())
				list.insert(&channelSpec{
					dataChannel:  dataCh,
					errorChannel: errCh,
					location:     location,
					firstValue:   firstPoint,
				})
			} else {
				fmt.Printf("# WARNING: no data for %s %s\n", server.Url, location.Location)
			}
		}
	}

	//
	// Run all trends as a merge sort
	//
	for item := list.popFirst(); item != nil; item = list.popFirst() {
		thisChannel := item.item
		value := thisChannel.firstValue

		output <- value

		// Fetch the next sample off of whatever one we pulled this data from
		nextValue := <-thisChannel.dataChannel
		eatError(thisChannel.errorChannel)

		// If the result of the last fetch was nil that means there is no more.
		// Otherwise, take this new sample and insert it into the list (merge sorted
		// to the corret place)
		if nextValue != nil {
			thisChannel.firstValue = nextValue
			list.insert(thisChannel)
		}
	}
	output <- nil
}

func eatError(errorChannel chan error) {
	select {
	case err := <-errorChannel:
		if err != nil {
			log.Fatal(err)
		}
	default:
		// do nothing
	}
}
