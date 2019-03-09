# esx

CLI for streaming I/O with Elasticsearch

## Installation


Navigate to the root `esx/` directory (where the `Makefile` is located) and run:

```
make install
```

## Usage

### `esx --help`

```
usage: esx --es-index=ES-INDEX [<flags>] <command> [<args> ...]

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
  -H, --es-host="localhost:9200"
                           ElasticSearch host:port
  -I, --es-index=ES-INDEX  ElasticSearch index to use
  -D, --es-type="_doc"     ElasticSearch doc type to use
  -T, --es-timeout=10s     ElasticSearch operation timeout duration
  -d, --debug              Debug mode
  -P, --progress           Report progress

Commands:
  help [<command>...]
    Show help.

  scroll [<flags>]
    Scrolls an ElasticSearch index

  index [<flags>]
    Indexes data into an ElasticSearch index
```

### `esx scroll --help`

```
usage: esx scroll [<flags>]

Scrolls an ElasticSearch index

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -H, --es-host="localhost:9200"
                               ElasticSearch host:port
  -I, --es-index=ES-INDEX      ElasticSearch index to use
  -D, --es-type="_doc"         ElasticSearch doc type to use
  -T, --es-timeout=10s         ElasticSearch operation timeout duration
  -d, --debug                  Debug mode
  -P, --progress               Report progress
  -s, --scroll-size=100        ElasticSearch scroll size
  -t, --scroll-timeout=10s     ElasticSearch scroll timeout duration
  -f, --query-file=QUERY-FILE  Query JSON file
  -q, --query=QUERY            Query string
```

### `esx index --help`

```
usage: esx index [<flags>]

Indexes data into an ElasticSearch index

Flags:
      --help                  Show context-sensitive help (also try --help-long and --help-man).
  -H, --es-host="localhost:9200"
                              ElasticSearch host:port
  -I, --es-index=ES-INDEX     ElasticSearch index to use
  -D, --es-type="_doc"        ElasticSearch doc type to use
  -T, --es-timeout=10s        ElasticSearch operation timeout duration
  -d, --debug                 Debug mode
  -P, --progress              Report progress
  -a, --index-action="index"  Index action type ("index" or "update")
  -w, --index-workers=0       Number of index workers
  -b, --batch-size=100        Number of documents to batch index
  -i, --doc-id-field="_id"    JSON field to use as document ID
```

## Contributing

When contributing to this repository, please follow the steps below:

1. Fork the repository
1. Submit your patch in one commit, or a series of well-defined commits
1. Submit your pull request and make sure you reference the issue you are addressing
