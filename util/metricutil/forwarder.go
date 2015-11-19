package metricutil

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mattbaird/elastigo/lib"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

var ActiveForwarder *elastigo.BulkIndexer

func StartEventForwarder(ctx context.Context) {
	url := os.Getenv("SG_ELASTICSEARCH_URL")
	if url == "" {
		log15.Error("EventForwarder failed to locate elasticsearch endpoint")
		return
	}

	conn := elastigo.NewConn()
	conn.SetFromUrl(url)

	ActiveForwarder = conn.NewBulkIndexerErrors(10, 60)
	if ActiveForwarder == nil {
		log15.Error("EventForwarder could not connect to elasticsearch")
		return
	}
	ActiveForwarder.Start()

	go func() {
		for errBuf := range ActiveForwarder.ErrorChannel {
			log15.Error("EventForwarder recieved error", "error", errBuf.Err)
		}
	}()

	log15.Debug("EventForwarder initialized")
}

func ForwardEvents(ctx context.Context, eventList *sourcegraph.UserEventList) {
	indexName := getIndexNameWithSuffix()
	if ActiveForwarder != nil {
		for _, event := range eventList.Events {
			indexNameWithPrefix := indexName
			if event.Version == "dev" {
				indexNameWithPrefix = "dev-" + indexNameWithPrefix
			}
			if err := ActiveForwarder.Index(indexNameWithPrefix, "user_event", "", "", "", nil, event); err != nil {
				log15.Error("EventForwarder failed to push event", "event", event, "error", err)
			}
		}
	}
}

func getIndexNameWithSuffix() string {
	t := time.Now().UTC()
	return fmt.Sprintf("events-%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}
