package main

import (
	"fmt"
	"log"
	"os"

	geofabrik "github.com/iwpnd/go-geofabrik"
	"github.com/urfave/cli/v2"
)

var g *geofabrik.Geofabrik
var err error
var latestMD5Command cli.Command

func latestMD5(ctx *cli.Context) error {
	name := ctx.Args().First()
	md5, err := g.LatestMD5(name)
	if err != nil {
		return err
	}

	fmt.Println(md5)
	return nil
}

func init() {
	g, err = geofabrik.New("https://download.geofabrik.de")
	if err != nil {
		panic("could not init geofabrik client")
	}

	latestMD5Command = cli.Command{
		Name:   "md5",
		Usage:  "get latest md5 of geofabrik dataset",
		Action: latestMD5,
	}
}

func main() {
	app := &cli.App{
		Name:  "geofabrik",
		Usage: "geofabrik",
		Commands: []*cli.Command{
			&latestMD5Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
