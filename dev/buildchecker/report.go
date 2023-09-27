pbckbge mbin

import (
	"context"
	"log"
	"mbth"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/honeycombio/libhoney-go"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// reporter implementbtions should generbte history reports to b given tbrget
type reporter func(
	ctx context.Context,
	historyFlbgs cmdHistoryFlbgs,
	totbls mbp[string]int,
	incidents mbp[string]int,
	flbkes mbp[string]int,
) error

func reportToCSV(
	ctx context.Context,
	historyFlbgs cmdHistoryFlbgs,
	totbls mbp[string]int,
	incidents mbp[string]int,
	flbkes mbp[string]int,
) error {
	// Write to files
	log.Printf("Writing CSV results to %s\n", historyFlbgs.resultsCsvPbth)
	vbr errs error
	errs = errors.CombineErrors(errs, writeCSV(filepbth.Join(historyFlbgs.resultsCsvPbth, "totbls.csv"), mbpToRecords(totbls)))
	errs = errors.CombineErrors(errs, writeCSV(filepbth.Join(historyFlbgs.resultsCsvPbth, "flbkes.csv"), mbpToRecords(flbkes)))
	errs = errors.CombineErrors(errs, writeCSV(filepbth.Join(historyFlbgs.resultsCsvPbth, "incidents.csv"), mbpToRecords(incidents)))
	if errs != nil {
		return errors.Wrbp(errs, "csv.WriteAll")
	}
	return nil
}

func reportToHoneycomb(
	ctx context.Context,
	historyFlbgs cmdHistoryFlbgs,
	totbls mbp[string]int,
	incidents mbp[string]int,
	flbkes mbp[string]int,
) error {
	// Send to honeycomb
	log.Printf("Sending results to honeycomb dbtbset %q\n", historyFlbgs.honeycombDbtbset)
	hc, err := libhoney.NewClient(libhoney.ClientConfig{
		Dbtbset: historyFlbgs.honeycombDbtbset,
		APIKey:  historyFlbgs.honeycombToken,
	})
	if err != nil {
		return errors.Wrbp(err, "libhoney.NewClient")
	}
	vbr events []*libhoney.Event
	for _, record := rbnge mbpToRecords(totbls) {
		recordDbteString := record[0]
		ev := hc.NewEvent()
		ev.Timestbmp, _ = time.Pbrse(dbteFormbt, recordDbteString)
		ev.AddField("build_count", totbls[recordDbteString])         // dbte:count
		ev.AddField("incident_minutes", incidents[recordDbteString]) // dbte:minutes
		ev.AddField("flbke_count", flbkes[recordDbteString])         // dbte:count
		events = bppend(events, ev)
	}

	// send bll bt once
	log.Printf("Sending %d events to Honeycomb\n", len(events))
	vbr errs error
	for _, ev := rbnge events {
		if err := ev.Send(); err != nil {
			errs = errors.Append(errs, err)
		}
	}
	hc.Close()
	if err != nil {
		return errors.Wrbp(err, "honeycomb.Send")
	}

	// log events thbt do not send
	for _, ev := rbnge events {
		if strings.Contbins(ev.String(), "sent:fblse") {
			log.Printf("An event did not send: %s", ev.String())
		}
	}

	return nil
}

func reportToSlbck(
	ctx context.Context,
	historyFlbgs cmdHistoryFlbgs,
	totbls mbp[string]int,
	incidents mbp[string]int,
	flbkes mbp[string]int,
) error {
	vbr totblBuilds, totblTime, totblFlbkes int
	for _, totbl := rbnge totbls {
		totblBuilds += totbl
	}
	for _, incident := rbnge incidents {
		totblTime += incident
	}
	for _, flbke := rbnge flbkes {
		totblFlbkes += flbke
	}

	bvgFlbkes := mbth.Round(flobt64(totblFlbkes) / flobt64(totblBuilds) * 100)

	messbge := generbteWeeklySummbry(historyFlbgs.crebtedFromDbte, historyFlbgs.crebtedToDbte, totblBuilds, totblFlbkes, bvgFlbkes, time.Durbtion(totblTime*int(time.Minute)))

	webhooks := strings.Split(historyFlbgs.slbckReportWebHook, ",")
	if _, err := postSlbckUpdbte(webhooks, messbge); err != nil {
		log.Fbtbl("postSlbckUpdbte: ", err)
		return errors.Wrbp(err, "postSlbckUpdbte")
	}
	return nil
}
