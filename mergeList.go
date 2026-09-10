package alcsoap

import (
	"fmt"
	"sync"
)

type SortedTrendList struct {
	head *listItem
	mu   sync.Mutex
}
type listItem struct {
	dataChannel <-chan TrendEvent
	event       TrendEvent
	next        *listItem
}

func NewSortedTrendList() *SortedTrendList {
	return &SortedTrendList{head: nil}
}

// insert() inserts an item into a singly linked list based on the timestamp.
func (list *SortedTrendList) insert(newEntry TrendEvent, source <-chan TrendEvent) *SortedTrendList {
	list.mu.Lock()
	defer list.mu.Unlock()

	newNode := &listItem{event: newEntry, dataChannel: source, next: nil}

	//fmt.Printf(">>> insert %s %s\n", newNode.event.Data.Time, newNode.event.Source.Location)
	if list.head == nil {
		// List is empty - this item is the head
		//fmt.Printf("!!! list was empty, this goes first\n")
		newNode.next = nil
		list.head = newNode
	} else if newNode.event.Data.Time.Before(list.head.event.Data.Time) {
		// This item goes before the first item
		newNode.next = list.head
		list.head = newNode
	} else {
		// Find where the new item goes and insert it then link to the current location
		var current = list.head
		for current.next != nil && newNode.event.Data.Time.After(current.next.event.Data.Time) {
			current = current.next
		}

		newNode.next = current.next
		current.next = newNode
	}

	//fmt.Print("I ")
	//list.dump()
	return list
}

func (list *SortedTrendList) popFirst() *listItem {
	list.mu.Lock()
	defer list.mu.Unlock()

	//fmt.Print("B ")
	//list.dump()

	item := list.head

	if item != nil {
		list.head = item.next
		//fmt.Printf("<<< pop    %s %s\n", item.event.Data.Time, item.event.Source.Location)
	}

	//fmt.Print("P ")
	//list.dump()

	return item
}

func (list *SortedTrendList) count() int {
	var count = 0

	for item := list.head; item != nil; item = item.next {
		count = count + 1
	}
	return count
}

func (list *SortedTrendList) dump() {
	fmt.Print("List: ")
	for item := list.head; item != nil; item = item.next {
		fmt.Printf("%s (%s) ->\n        ", item.event.Data.Time, item.event.Source.Location)
	}
	fmt.Println("nil")
}
