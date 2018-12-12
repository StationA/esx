package main

import (
	"fmt"
	"github.com/olivere/elastic"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"runtime"
)

const WorkersPerCPU = 2

var (
	esHost        = kingpin.Flag("es-host", "ElasticSearch host:port").Short('H').Envar("ES_HOST").Default("localhost:9200").String()
	esIndex       = kingpin.Flag("es-index", "ElasticSearch index to use").Short('I').Envar("ES_INDEX").Required().String()
	esType        = kingpin.Flag("es-type", "ElasticSearch doc type to use").Short('D').Envar("ES_TYPE").Default("_doc").String()
	esTimeout     = kingpin.Flag("es-timeout", "ElasticSearch operation timeout duration").Short('T').Default("10s").Duration()
	debug         = kingpin.Flag("debug", "Debug mode").Short('d').Bool()
	progress      = kingpin.Flag("progress", "Report progress").Short('P').Bool()
	scrollCmd     = kingpin.Command("scroll", "Scrolls an ElasticSearch index")
	scrollSize    = scrollCmd.Flag("scroll-size", "ElasticSearch scroll size").Short('s').Default("100").Int()
	scrollTimeout = scrollCmd.Flag("scroll-timeout", "ElasticSearch scroll timeout duration").Short('t').Default("10s").Duration()
	queryFile     = scrollCmd.Flag("query-file", "Query JSON file").Short('f').File()
	queryStr      = scrollCmd.Flag("query", "Query string").Short('q').String()
	indexCmd      = kingpin.Command("index", "Indexes data into an ElasticSearch index")
	indexAction   = indexCmd.Flag("index-action", "Index action type (\"index\" or \"update\")").Short('a').Default("update").String()
	numWorkers    = indexCmd.Flag("index-workers", "Number of index workers").Short('w').Default("0").Int()
	batchSize     = indexCmd.Flag("batch-size", "Number of documents to batch index").Short('b').Default("100").Int()
	docIdField    = indexCmd.Flag("doc-id-field", "JSON field to use as document ID").Short('i').Default("_id").String()
	ProgressBar   = &Progress{}
)

func handleErr(cmd string, err error) {
	if err != nil {
		fmt.Printf("%s failed: %v\n", cmd, err)
		os.Exit(1)
	}
}

func main() {
	cmd := kingpin.Parse()

	numCores := runtime.NumCPU()
	if *numWorkers < 1 {
		*numWorkers = WorkersPerCPU * numCores
	}
	runtime.GOMAXPROCS(numCores)

	if *progress {
		ProgressBar.Enable()
	}

	client, err := elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s", *esHost)),
		elastic.SetGzip(true),
	)
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	switch cmd {
	case "scroll":
		handleErr(cmd, doScroll(client))
	case "index":
		handleErr(cmd, doIndex(client))
	}
}
