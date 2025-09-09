package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	geofabrik "github.com/iwpnd/go-geofabrik"
	"github.com/iwpnd/rip"
	"github.com/urfave/cli/v3"
)

var (
	g                        *geofabrik.Geofabrik
	err                      error
	latestMD5Command         cli.Command
	polygonCommand           cli.Command
	downloadCommand          cli.Command
	downloadIfChangedCommand cli.Command
)

var (
	md5Flag        cli.StringFlag
	outputPathFlag cli.StringFlag
)

func latestMD5(ctx context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	md5, err := g.MD5(ctx, name)
	if err != nil {
		return err
	}

	fmt.Println(md5)
	return nil
}

func polygon(ctx context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	polygon, err := g.Polygon(ctx, name)
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

func downloadIfChanged(ctx context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()

	latestMD5, err := g.MD5(ctx, name)
	if err != nil {
		return err
	}
	md5 := cmd.String("md5")
	if md5 == latestMD5 {
		fmt.Printf(
			"%s is up to date, no download required (latest md5: %s, input md5: %s)\n\n",
			name,
			latestMD5,
			md5,
		)
		return nil
	}

	outputPath := cmd.String("outputPath")

	fmt.Printf("downloading %s (%s) \n\n", name, latestMD5)
	err = g.Download(ctx, name, outputPath)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Printf("download canceled for %s (%s) \n\n", name, latestMD5)
			return err
		}
		fmt.Printf("download error: %s \n\n", err)
		return err
	}
	fmt.Printf("\n\nfinished downloading %s (%s)", name, latestMD5)

	return nil
}

func download(ctx context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()

	outputPath := cmd.String("outputPath")

	fmt.Printf("downloading %s \n\n", name)
	err = g.Download(ctx, name, outputPath)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Printf("download canceled for %s \n\n", name)
			return err
		}
		fmt.Printf("download error: %s \n\n", err)
		return err
	}
	fmt.Printf("\n\nfinished downloading %s", name)

	return nil
}

func init() {
	g, err = geofabrik.New(
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
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	go func() {
		<-sigCh
		fmt.Println("shutdown requested; cleaning up…")

		go func() {
			<-sigCh
			fmt.Println("forced exit")
			os.Exit(1)
		}()

		// after 5s, cancel the shutdownCtx
		time.Sleep(5 * time.Second)
		shutdownCancel()
	}()

	app := &cli.Command{
		Name:  "geofabrik",
		Usage: "geofabrik",
		Commands: []*cli.Command{
			&latestMD5Command,
			&polygonCommand,
			&downloadCommand,
			&downloadIfChangedCommand,
		},
	}

	if err := app.Run(shutdownCtx, os.Args); err != nil {
		os.Exit(1) //nolint:gocritic
	}
}
