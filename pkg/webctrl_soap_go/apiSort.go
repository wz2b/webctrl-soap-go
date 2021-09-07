package webctrl_soap_go

func SortGroup(input []TrendEvent, order []*TrendSource) []*TrendEvent {

	outputGroup := make([](*TrendEvent), len(order))

	var temp = make(map[TrendSource](TrendEvent), len(input))
	for _, event := range input {
		key := *event.Source
		temp[key] = event
	}

	for i, source := range order {
		key := source
		entry := temp[*key]
		outputGroup[i] = &entry
	}

	return outputGroup
}
