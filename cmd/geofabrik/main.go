package main

import (
	"fmt"
	"log"
	"os"

	geofabrik "github.com/iwpnd/go-geofabrik"
	"github.com/iwpnd/rip"
	"github.com/urfave/cli/v2"
)

var g *geofabrik.Geofabrik
var err error
var latestMD5Command cli.Command
var simpleDownloadCommand cli.Command

func latestMD5(ctx *cli.Context) error {
	name := ctx.Args().First()
	md5, err := g.LatestMD5(name)
	if err != nil {
		return err
	}

	fmt.Println(md5)
	return nil
}

func simpleDownload(ctx *cli.Context) error {
	name := ctx.Args().First()

	fmt.Printf("downloading '%s' \n\n", name)
	err := g.SimpleDownload(name, ".")
	if err != nil {
		return err
	}
	fmt.Printf("\n\nfinished downloading '%s'", name)

	return nil
}

func init() {
	g, err = geofabrik.NewWithProgress(
		"https://download.geofabrik.de",
		rip.WithTimeout(0),
	)
	if err != nil {
		panic("could not init geofabrik client")
	}

	latestMD5Command = cli.Command{
		Name:   "md5",
		Usage:  "get latest md5 of geofabrik dataset",
		Action: latestMD5,
	}
	simpleDownloadCommand = cli.Command{
		Name:   "download",
		Usage:  "download dataset to outputpath",
		Action: simpleDownload,
	}
}

func main() {
	app := &cli.App{
		Name:  "geofabrik",
		Usage: "geofabrik",
		Commands: []*cli.Command{
			&latestMD5Command,
			&simpleDownloadCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
