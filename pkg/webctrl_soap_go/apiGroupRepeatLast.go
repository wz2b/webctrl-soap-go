package webctrl_soap_go

func GroupRepeatLast(input <-chan TrendGroupEvent) <-chan TrendGroupEvent {

	output := make(chan TrendGroupEvent)

	go func() {

		last := make(map[TrendSource]TrendEvent)

		for ingrp := range input {
			if ingrp.IsEmpty() {
				break
			}

			outgrp := make(map[TrendSource]TrendEvent)

			// All the new fields will update last
			for key, entry := range ingrp.Trends {
				if entry.Data.ValueString != "" {
					last[key] = entry
				}
			}

			// Now, everything in 'last' is what we want to output
			for key, entry := range last {
				outgrp[key] = entry
			}

			output <- TrendGroupEvent{Time: ingrp.Time, Trends: outgrp}
		}

		close(output)
	}()

	return output
}
