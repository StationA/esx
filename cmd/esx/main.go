package main

import (
	"fmt"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"runtime"
)

const WorkersPerCPU = 2

var (
	esHost             = kingpin.Flag("es-host", "ElasticSearch host:port").Short('H').Envar("ES_HOST").Default("localhost:9200").String()
	esIndex            = kingpin.Flag("es-index", "ElasticSearch index to use").Short('I').Envar("ES_INDEX").Required().String()
	esType             = kingpin.Flag("es-type", "ElasticSearch doc type to use").Short('D').Envar("ES_TYPE").Default("_doc").String()
	esTimeout          = kingpin.Flag("es-timeout", "ElasticSearch operation timeout duration").Short('T').Default("60s").Duration()
	debug              = kingpin.Flag("debug", "Debug mode").Short('d').Bool()
	quiet              = kingpin.Flag("quiet", "Silences all log output").Short('q').Bool()
	progress           = kingpin.Flag("progress", "Report progress").Short('P').Bool()
	scrollCmd          = kingpin.Command("scroll", "Scrolls an ElasticSearch index")
	scrollSize         = scrollCmd.Flag("scroll-size", "ElasticSearch scroll size").Short('s').Default("100").Int()
	scrollTimeout      = scrollCmd.Flag("scroll-timeout", "ElasticSearch scroll timeout duration").Short('t').Default("10s").Duration()
	queryFile          = scrollCmd.Flag("query-file", "Query JSON file").Short('f').File()
	queryStr           = scrollCmd.Flag("query", "Query string").Short('Q').String()
	indexCmd           = kingpin.Command("index", "Indexes data into an ElasticSearch index")
	indexAction        = indexCmd.Flag("index-action", "Index action type (\"index\" or \"update\")").Short('a').Default("index").String()
	docIdField         = indexCmd.Flag("doc-id-field", "JSON field to use as document ID").Short('i').Default("_id").String()
	numWorkers         = indexCmd.Flag("index-workers", "Number of index workers").Short('w').Default("0").Int()
	batchSize          = indexCmd.Flag("batch-size", "Number of documents to batch index").Short('b').Default("100").Int()
	numRetries         = indexCmd.Flag("num-retries", "Number of times to retry a failed batch").Short('r').Default("3").Int()
	throttleHWM        = indexCmd.Flag("throttle-high-water-mark", "ADVANCED: Percentage of limit to consider as high water mark").Short('M').Default("75").Int()
	throttleWindowSize = indexCmd.Flag("throttle-window-size", "ADVANCED: Number of runtime samples to use for estimating index timing").Short('W').Default("50").Int()
	queueFullWait      = indexCmd.Flag("queue-full-wait-time", "ADVANCED: How long to wait for the work queue to free up").Short('u').Default("10s").Duration()
	ProgressBar        = &Progress{}
	Log                = logrus.New()
)

func handleErr(cmd string, err error) {
	if err != nil {
		Log.Errorf("%s failed: %v", cmd, err)
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

	clientOpts := []elastic.ClientOptionFunc{
		elastic.SetURL(fmt.Sprintf("http://%s", *esHost)),
		elastic.SetGzip(true),
	}

	Log.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	if *quiet {
		Log.SetOutput(ioutil.Discard)
	} else {
		Log.SetOutput(os.Stderr)
	}
	if *debug {
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}

	client, err := elastic.NewClient(clientOpts...)
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
