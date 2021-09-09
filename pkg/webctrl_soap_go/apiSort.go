package webctrl_soap_go

func SortGroup(inputGroup TrendGroupEvent, order []TrendSource) OrderedTrendGroupEvent {

	outputGroup := OrderedTrendGroupEvent{Time: inputGroup.Time, Trends: make([]TrendEvent, len(order))}

	for i, source := range order {
		entry, ok := inputGroup.Trends[source]
		if ok {
			outputGroup.Trends[i] = entry

		}
	}

	return outputGroup
}
