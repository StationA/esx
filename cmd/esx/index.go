package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
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

func indexBatch(ctx context.Context, client *elastic.Client, batch Batch) (time.Duration, error) {
	indexCtx, indexCancel := context.WithTimeout(ctx, *esTimeout)
	defer indexCancel()
	log := Log.
		WithField("proc", ctx.Value("worker-id")).
		WithField("batch", batch.ID)

	start := time.Now()
	bulk := client.Bulk()
	for _, doc := range batch.Docs {
		docId := doc[*docIdField]
		if docId == nil {
			return 0, fmt.Errorf("Document ID field [%s] is not set on document: %+v", *docIdField, doc)
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
	res, err := bulk.Do(indexCtx)
	duration := time.Since(start)
	if err != nil {
		return 0, err
	}
	failed := res.Failed()
	if len(failed) > 0 {
		for _, failure := range failed {
			Log.
				WithField("doc-id", failure.Id).
				Errorf("Document failed to index: %+v", failure.Error)
		}
		return 0, fmt.Errorf("Batch [%d] failed: %+v", batch.ID, failed)
	} else {
		log.Infof("Batch completed in %.2fs", duration.Seconds())
	}
	return duration, nil
}

func indexWorker(ctx context.Context, client *elastic.Client, batches <-chan Batch) error {
	throttle := ctx.Value("throttle").(*SamplingThrottle)
	log := Log.WithField("proc", ctx.Value("worker-id"))
	for batch := range batches {
		select {
		case <-ctx.Done():
			log.Warnf("Context cancelled; shutting down")
			return nil
		case <-throttle.Wait():
			retry := 0
			for {
				t, err := indexBatch(ctx, client, batch)
				throttle.Collect(t)
				if err != nil {
					if retry < *numRetries {
						retry += 1
						log.Warnf("Batch failed, retrying (%d of %d)", retry, *numRetries)
					} else {
						return err
					}
				} else {
					break
				}
			}
		}
	}
	return nil
}

func producer(ctx context.Context, client *elastic.Client, batches chan<- Batch) error {
	defer close(batches)

	var batch Batch
	log := Log.WithField("proc", "producer").WithField("batch", batch.ID)
	dec := json.NewDecoder(os.Stdin)
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
			shouldRetry := true
			for shouldRetry {
				log.Debugf("Enqueuing batch")
				select {
				case batches <- batch:
					log.Debugf("Batch enqueued")
					batch = Batch{ID: batch.ID + 1}
					log = log.WithField("batch", batch.ID)
					shouldRetry = false
				default:
					log.Warnf("Work queue is full; retrying in %.2fs", queueFullWait.Seconds())
					select {
					case <-ctx.Done():
						log.Warnf("Context cancelled; shutting down")
						return nil
					case <-time.After(*queueFullWait):
					}
				}
			}
		}
	}
	// Flush any remaining documents
	if len(batch.Docs) > 0 {
		batches <- batch
	}
	return nil
}

func doIndex(client *elastic.Client) error {
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batches := make(chan Batch, (*numWorkers)*2)
	throttle := NewSamplingThrottle(
		SetLimit(*esTimeout),
		SetWindowSize(50),
	)

	g, ctx := errgroup.WithContext(cancelCtx)
	workersCtx := context.WithValue(ctx, "throttle", throttle)
	for i := 0; i < *numWorkers; i++ {
		worker := fmt.Sprintf("worker-%d", i)
		g.Go(func() error {
			Log.Infof("%s started", worker)
			workerCtx := context.WithValue(workersCtx, "worker-id", worker)
			err := indexWorker(workerCtx, client, batches)
			if err != nil {
				Log.Errorf("%s failed: %v", worker, err)
				cancel()
			}
			return err
		})
	}
	g.Go(func() error {
		Log.Infof("producer started")
		err := producer(ctx, client, batches)
		if err != nil {
			Log.WithError(err).Errorf("producer failed")
			cancel()
		}
		return err
	})
	return g.Wait()
}
