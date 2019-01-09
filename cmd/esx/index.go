package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"strings"
	"time"
)

type Batch struct {
	ID   int
	Docs []map[string]interface{}
}

func indexWorker(ctx context.Context, client *elastic.Client, batches <-chan Batch) error {
	for batch := range batches {
		start := time.Now()
		bulk := client.Bulk()
		for _, doc := range batch.Docs {
			docId := doc[*docIdField]
			if docId == nil {
				return fmt.Errorf("Document ID field [%s] is not set on document: %+v", *docIdField, doc)
			}
			for k, _ := range doc {
				// Ignore all fields that start with an underscore
				if strings.HasPrefix(k, "_") {
					delete(doc, k)
				}
			}
			req := elastic.NewBulkIndexRequest().
				OpType(*indexAction).
				Index(*esIndex).
				Type(*esType).
				Id(fmt.Sprintf("%v", docId)).
				Doc(doc)
			bulk.Add(req)
		}
		res, err := bulk.Do(ctx)
		duration := time.Since(start)
		log.Infof("Batch #%d (%d docs) took %.2fs to index\n", batch.ID, len(batch.Docs), duration.Seconds())
		if err != nil {
			return err
		}
		failed := res.Failed()
		if len(failed) > 0 {
			for _, failure := range failed {
				fmt.Fprintf(os.Stderr, "%s: %v\n", failure.Id, failure.Error)
			}
			return fmt.Errorf("Batch failed: %+v", failed)
		}
	}
	return nil
}

func producer(ctx context.Context, client *elastic.Client, batches chan<- Batch) error {
	dec := json.NewDecoder(os.Stdin)
	var batch Batch
	for {
		var doc map[string]interface{}
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		batch.Docs = append(batch.Docs, doc)
		if len(batch.Docs) == *batchSize {
			batches <- batch
			batch = Batch{ID: batch.ID + 1}
		}
	}
	// Flush any remaining documents
	if len(batch.Docs) > 0 {
		batches <- batch
	}
	close(batches)
	return nil
}

func doIndex(client *elastic.Client) error {
	g, ctx := errgroup.WithContext(context.Background())
	batches := make(chan Batch, (*numWorkers)*2)
	for i := 0; i < *numWorkers; i++ {
		g.Go(func() error {
			return indexWorker(ctx, client, batches)
		})
	}
	g.Go(func() error {
		return producer(ctx, client, batches)
	})
	return g.Wait()
}
