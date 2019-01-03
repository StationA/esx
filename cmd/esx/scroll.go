package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/olivere/elastic"
	"io"
	"os"
)

func doScroll(client *elastic.Client) error {
	ctx := context.Background()
	enc := json.NewEncoder(os.Stdout)
	queryFile := *queryFile
	queryStr := *queryStr
	query := "{}"
	if (queryStr != "" && queryFile != nil) {
		return errors.New("Scroll cannot accept a query file and a query string.")
	} else if queryStr != "" {
		query = queryStr
	} else if queryFile != nil {
		data := make([]byte, 5000)
		_, err := queryFile.Read(data)
		if err != nil {
			return err
		}
		query = string(data)
	}
	scroll := client.Scroll(*esIndex).Size(*scrollSize).Body(query)

	if ProgressBar.IsEnabled() {
		countCtx, countCancel := context.WithTimeout(ctx, *esTimeout)
		defer countCancel()

		total, err := client.Count(*esIndex).Do(countCtx)
		if err != nil {
			return err
		}

		ProgressBar.SetTotal(int(total))
	}

	for {
		scrollCtx, scrollCancel := context.WithTimeout(ctx, *scrollTimeout)
		results, err := scroll.Do(scrollCtx)
		if err == io.EOF {
			scrollCancel()
			break
		}
		if err != nil {
			scrollCancel()
			return err
		}
		for _, hit := range results.Hits.Hits {
			var source map[string]interface{}
			err := json.Unmarshal(*hit.Source, &source)
			if err != nil {
				return err
			}
			source["_index"] = hit.Index
			source["_type"] = hit.Type
			source["_id"] = hit.Id
			err = enc.Encode(source)
			if err != nil {
				return err
			}
			ProgressBar.Increment()
		}
		scrollCancel()
	}
	return nil
}
