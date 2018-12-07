package main

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic"
	"io"
	"os"
)

func doScroll(client *elastic.Client) error {
	ctx := context.Background()
	enc := json.NewEncoder(os.Stdout)
	scroll := client.Scroll(*esIndex).Size(*scrollSize)

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
