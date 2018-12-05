package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"io"
	"os"
)

func doIndex(client *elastic.Client) error {
	dec := json.NewDecoder(os.Stdin)
	var batch = client.Bulk()
	for {
		var doc map[string]interface{}
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		docId := doc["id"].(string)
		delete(doc, "id")
		req := elastic.NewBulkIndexRequest().
			Type("doc_as_upsert").
			Index(*esIndex).
			Type(*esType).
			Id(docId).
			Doc(doc)
		batch.Add(req)
		if batch.NumberOfActions() == *batchSize {
			res, err := batch.Do(context.Background())
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
			batch = client.Bulk()
		}
	}
	// Flush any remaining documents
	if batch.NumberOfActions() > 0 {
		res, err := batch.Do(context.Background())
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
