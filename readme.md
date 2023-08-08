# go-geofabrik

unofficial api client for [Geofabrik](https://download.geofabrik.de).

## Installation

### cli

```bash
go install github.com/iwpnd/go-geofabrik/cmd/geofabrik@latest
```

```bash
➜ geofabrik --help
NAME:
   geofabrik - geofabrik

USAGE:
   geofabrik [global options] command [command options] [arguments...]

COMMANDS:
   md5                  get latest md5 of geofabrik dataset
   polygon              get extent of dataset as geojson feature
   download             download dataset to outputpath
   download-if-changed  download dataset to outputpath if md5 changed
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

## License

MIT

## Acknowledgement

awesome folks at [Geofabrik](https://geofabrik.de)

## Maintainer

Benjamin Ramser - [@iwpnd](https://github.com/iwpnd)

Project Link: [https://github.com/iwpnd/go-geofabrik](https://github.com/iwpnd/go-geofabrik)
