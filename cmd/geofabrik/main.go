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
var polygonCommand cli.Command
var downloadCommand cli.Command
var downloadIfChangedCommand cli.Command

var md5Flag cli.StringFlag
var outputPathFlag cli.StringFlag

func latestMD5(ctx *cli.Context) error {
	name := ctx.Args().First()
	md5, err := g.MD5(name)
	if err != nil {
		return err
	}

	fmt.Println(md5)
	return nil
}

func polygon(ctx *cli.Context) error {
	name := ctx.Args().First()
	polygon, err := g.Polygon(name)
	if err != nil {
		return err
	}
	f, err := polygon.ToFeature()
	if err != nil {
		return err
	}

	fmt.Println(f)
	return nil
}

func downloadIfChanged(ctx *cli.Context) error {
	name := ctx.Args().First()

	latestMD5, err := g.MD5(name)
	if err != nil {
		return err
	}
	md5 := ctx.String("md5")
	if md5 == latestMD5 {
		fmt.Printf(
			"%s is up to date, no download required (latest md5: %s, input md5: %s)\n\n",
			name,
			latestMD5,
			md5,
		)
		return nil
	}

	outputPath := ctx.String("outputPath")

	fmt.Printf("downloading %s (%s) \n\n", name, latestMD5)
	err = g.Download(name, outputPath)
	if err != nil {
		return err
	}
	fmt.Printf("\n\nfinished downloading %s (%s)", name, latestMD5)

	return nil
}

func download(ctx *cli.Context) error {
	name := ctx.Args().First()

	outputPath := ctx.String("outputPath")

	fmt.Printf("downloading %s \n\n", name)
	err = g.Download(name, outputPath)
	if err != nil {
		return err
	}
	fmt.Printf("\n\nfinished downloading %s", name)

	return nil
}

func init() {
	g, err = geofabrik.NewWithProgress(
		"http://download.geofabrik.de",
		rip.WithTimeout(0),
	)
	if err != nil {
		panic("could not init geofabrik client")
	}

	md5Flag = cli.StringFlag{
		Name:     "md5",
		Required: true,
		Usage:    "md5 to compare",
	}
	outputPathFlag = cli.StringFlag{
		Name:     "outputPath",
		Required: true,
		Usage:    "path to store dataset",
	}

	latestMD5Command = cli.Command{
		Name:   "md5",
		Usage:  "get latest md5 of geofabrik dataset",
		Action: latestMD5,
	}
	polygonCommand = cli.Command{
		Name:   "polygon",
		Usage:  "get extent of dataset as geojson feature",
		Action: polygon,
	}
	downloadCommand = cli.Command{
		Name:   "download",
		Usage:  "download dataset to outputpath",
		Action: download,
		Flags: []cli.Flag{
			&outputPathFlag,
		},
	}
	downloadIfChangedCommand = cli.Command{
		Name:   "download-if-changed",
		Usage:  "download dataset to outputpath if md5 changed",
		Action: downloadIfChanged,
		Flags: []cli.Flag{
			&md5Flag,
			&outputPathFlag,
		},
	}
}

func main() {
	app := &cli.App{
		Name:  "geofabrik",
		Usage: "geofabrik",
		Commands: []*cli.Command{
			&latestMD5Command,
			&polygonCommand,
			&downloadCommand,
			&downloadIfChangedCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
