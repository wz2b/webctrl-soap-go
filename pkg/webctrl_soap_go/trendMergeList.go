package webctrl_soap_go

type sortedList struct {
	head *listItem
}
type listItem struct {
	dataChannel <-chan *TrendEvent
	event       *TrendEvent
	next        *listItem
}

// insert() inserts an item into a singly linked list based on the timestamp.
func (list *sortedList) insert(newEntry *TrendEvent, source <-chan *TrendEvent) *sortedList {
	newNode := &listItem{event: newEntry, dataChannel: source, next: nil}

	if list.head == nil {
		// List is empty - this item is the head
		newNode.next = nil
		list.head = newNode
	} else if newEntry.Data.Time.Before(*list.head.event.Data.Time) {
		// This item goes before the first item
		newNode.next = list.head
		list.head = newNode
	} else {
		// Find where the new item goes and insert it then link to the current location
		var current = list.head
		for current.next != nil && newEntry.Data.Time.After(*current.event.Data.Time) {
			current = current.next
		}

		newNode.next = current.next
		current.next = newNode
	}

	return list
}

func (list *sortedList) popFirst() *listItem {
	item := list.head

	if item != nil {
		list.head = item.next
	}

	return item
}
