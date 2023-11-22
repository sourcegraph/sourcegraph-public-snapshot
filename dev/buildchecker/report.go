package main

import (
	"context"
	"log"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/honeycombio/libhoney-go"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// reporter implementations should generate history reports to a given target
type reporter func(
	ctx context.Context,
	historyFlags cmdHistoryFlags,
	totals map[string]int,
	incidents map[string]int,
	flakes map[string]int,
) error

func reportToCSV(
	ctx context.Context,
	historyFlags cmdHistoryFlags,
	totals map[string]int,
	incidents map[string]int,
	flakes map[string]int,
) error {
	// Write to files
	log.Printf("Writing CSV results to %s\n", historyFlags.resultsCsvPath)
	var errs error
	errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "totals.csv"), mapToRecords(totals)))
	errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "flakes.csv"), mapToRecords(flakes)))
	errs = errors.CombineErrors(errs, writeCSV(filepath.Join(historyFlags.resultsCsvPath, "incidents.csv"), mapToRecords(incidents)))
	if errs != nil {
		return errors.Wrap(errs, "csv.WriteAll")
	}
	return nil
}

func reportToHoneycomb(
	ctx context.Context,
	historyFlags cmdHistoryFlags,
	totals map[string]int,
	incidents map[string]int,
	flakes map[string]int,
) error {
	// Send to honeycomb
	log.Printf("Sending results to honeycomb dataset %q\n", historyFlags.honeycombDataset)
	hc, err := libhoney.NewClient(libhoney.ClientConfig{
		Dataset: historyFlags.honeycombDataset,
		APIKey:  historyFlags.honeycombToken,
	})
	if err != nil {
		return errors.Wrap(err, "libhoney.NewClient")
	}
	var events []*libhoney.Event
	for _, record := range mapToRecords(totals) {
		recordDateString := record[0]
		ev := hc.NewEvent()
		ev.Timestamp, _ = time.Parse(dateFormat, recordDateString)
		ev.AddField("build_count", totals[recordDateString])         // date:count
		ev.AddField("incident_minutes", incidents[recordDateString]) // date:minutes
		ev.AddField("flake_count", flakes[recordDateString])         // date:count
		events = append(events, ev)
	}

	// send all at once
	log.Printf("Sending %d events to Honeycomb\n", len(events))
	var errs error
	for _, ev := range events {
		if err := ev.Send(); err != nil {
			errs = errors.Append(errs, err)
		}
	}
	hc.Close()
	if err != nil {
		return errors.Wrap(err, "honeycomb.Send")
	}

	// log events that do not send
	for _, ev := range events {
		if strings.Contains(ev.String(), "sent:false") {
			log.Printf("An event did not send: %s", ev.String())
		}
	}

	return nil
}

func reportToSlack(
	ctx context.Context,
	historyFlags cmdHistoryFlags,
	totals map[string]int,
	incidents map[string]int,
	flakes map[string]int,
) error {
	var totalBuilds, totalTime, totalFlakes int
	for _, total := range totals {
		totalBuilds += total
	}
	for _, incident := range incidents {
		totalTime += incident
	}
	for _, flake := range flakes {
		totalFlakes += flake
	}

	avgFlakes := math.Round(float64(totalFlakes) / float64(totalBuilds) * 100)

	message := generateWeeklySummary(historyFlags.createdFromDate, historyFlags.createdToDate, totalBuilds, totalFlakes, avgFlakes, time.Duration(totalTime*int(time.Minute)))

	webhooks := strings.Split(historyFlags.slackReportWebHook, ",")
	if _, err := postSlackUpdate(webhooks, message); err != nil {
		log.Fatal("postSlackUpdate: ", err)
		return errors.Wrap(err, "postSlackUpdate")
	}
	return nil
}
