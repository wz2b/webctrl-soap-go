package webctrl_soap_go

import (
	"fmt"
	"log"
	"time"
)

type channelSpec struct {
	dataChannel  chan *TrendPoint
	errorChannel chan error
	gql          string
	firstValue   *TrendPoint
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
func (list *sortedList) insert(item *channelSpec) *sortedList {
	newNode := &listItem{item: item, next: nil}
	// Special case for inserting this item at the beginning of the list
	if list.head == nil || item.firstValue.Time.Before(*list.head.item.firstValue.Time) {
		list.head = &listItem{item: item, next: nil}
	} else {
		var current *listItem = list.head
		for current.next != nil && item.firstValue.Time.After(*current.item.firstValue.Time) {
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

func (this *TrendService) MergeTendData(gqlPaths []string, startTime time.Time, endTime time.Time) {
	var list = &sortedList{head: nil}

	//
	// Populate the list with the first value of every trend
	//
	for _, gql := range gqlPaths {
		dataChannel, errorChannel := this.GetTrendData(gql, startTime, endTime)

		// get first data point
		firstPoint := <-dataChannel

		// non-blocking check for errors
		eatError(errorChannel)

		if firstPoint != nil {
			//fmt.Printf("Insert %s\n", firstPoint.Time.Local())
			list.insert(&channelSpec{
				dataChannel:  dataChannel,
				errorChannel: errorChannel,
				gql:          gql,
				firstValue:   firstPoint,
			})
		}
	}

	for item := list.popFirst(); item != nil; item = list.popFirst() {
		thisChannel := item.item
		value := thisChannel.firstValue

		fmt.Printf("%s %s %f\n", thisChannel.gql, value.Time.Local(), value.Value)

		nextValue := <-thisChannel.dataChannel
		//fmt.Printf("Pop out %s\n", nextValue.Time.Local())
		eatError(thisChannel.errorChannel)

		if nextValue != nil {
			thisChannel.firstValue = nextValue
			list.insert(thisChannel)
		}
	}
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
