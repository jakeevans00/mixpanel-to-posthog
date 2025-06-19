package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/posthog/posthog-go"
)

func PosthogImport(client posthog.Client, data []MixpanelDataLine) error {
	for _, line := range data {
		// Map the event name
		if line.Event == "$mp_web_page_view" {
			line.Event = "$pageview"
		} else {
			line.Event = MapEventName(line.Event, line.Properties)
		}

		// Construct properties
		properties := posthog.NewProperties()
		for k, v := range line.Properties {
			properties.Set(k, v)
		}
		properties.Set("$geoip_disable", true)
		properties.Set("$go_flag", "prod-import-1")
		err := client.Enqueue(posthog.Capture{
			DistinctId: line.DistinctID,
			Event:      line.Event,
			Properties: properties,
			Timestamp:  line.Time,
		})
		if err != nil {
			color.Red("\nError importing event: %s", line.Event)
			return err
		}
		// Sleep in between to avoid overloading the API
		time.Sleep(DELAY_MS * time.Millisecond)
	}
	return nil
}
