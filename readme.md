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

### package

### MD5

Get latest md5 of a dataset by name

```go
import (
    "contxt"
    "fmt"

    "github.com/iwpnd/go-geofabrik"
    "github.com/iwpnd/rip"
)

func main() {
    g, err := geofabrik.New("http://download.geofabrik.de", false)
    if err != nil {
        panic("wuaah!")
    }

    ctx := context.Background()
    name := "europe/germany/berlin"
    md5, err := g.MD5(ctx, name)
    if err != nil {
        panic(err)
    }
    fmt.Println(md5)
    // >> 379b462358f660744c1a9eed6f46b031
}
```

### Download

Get a dataset by name to output path

```go
import (
    "contxt"
    "fmt"

    "github.com/iwpnd/go-geofabrik"
    "github.com/iwpnd/rip"
)

func main() {
    g, err := geofabrik.New("http://download.geofabrik.de", false)
    if err != nil {
        panic("wuaah!")
    }

    ctx := context.Background()
    name := "europe/germany/berlin"
    outputPath := "./tmp"
    err := g.Download(ctx, name, outputPath)
    if err != nil {
        panic(err)
    }
}
```

### Polygon

Get a dataset extend as Polygon Feature

```go
import (
    "contxt"
    "fmt"

    "github.com/iwpnd/go-geofabrik"
    "github.com/iwpnd/rip"
)

func main() {
    g, err := geofabrik.New("http://download.geofabrik.de", false)
    if err != nil {
        panic("wuaah!")
    }

    ctx := context.Background()
    name := "europe/germany/berlin"
    polygon, err := g.Polygon(ctx, name)
    if err != nil {
        panic(err)
    }
    f, err := polygon.ToFeature()
    if err != nil {
        panic(err)
    }

    fmt.Println(f)
    // > {"type":"Feature"...}
}
```

## License

MIT

## Acknowledgement

awesome folks at [Geofabrik](https://geofabrik.de)

## Maintainer

Benjamin Ramser - [@iwpnd](https://github.com/iwpnd)

Project Link: [https://github.com/iwpnd/go-geofabrik](https://github.com/iwpnd/go-geofabrik)
