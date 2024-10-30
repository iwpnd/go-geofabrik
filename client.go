package geofabrik

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iwpnd/rip"
)

type FileType string

const (
	md5type  FileType = ".osm.pbf.md5"
	pbftype  FileType = ".osm.pbf"
	polytype FileType = ".poly"
)

// Geofabrik ...
type Geofabrik struct {
	*rip.Client
	withProgress bool
	progress     *Progress
}

// New ...
func New(host string, options ...rip.Option) (*Geofabrik, error) {
	c, err := rip.NewClient(host, options...)
	if err != nil {
		return &Geofabrik{}, err
	}

	return &Geofabrik{
		Client:       c,
		withProgress: false,
		progress:     nil,
	}, nil
}

// NewWithProgress will return a client including progress bar
func NewWithProgress(host string, options ...rip.Option) (*Geofabrik, error) {
	c, err := rip.NewClient(host, options...)
	if err != nil {
		return &Geofabrik{}, err
	}

	return &Geofabrik{
		Client:       c,
		withProgress: true,
		progress:     newProgress(),
	}, nil
}

// MD5 will return the latest MD5 of a dataset
func (g *Geofabrik) MD5(ctx context.Context, name string) (string, error) {
	p, err := newPath(name, md5type)
	if err != nil {
		return "", err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return "", ErrDownloadFailed{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		}
	}

	if res.StatusCode() >= 400 {
		return "", ErrDownloadFailed{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		}
	}
	defer res.Close()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}

func (g *Geofabrik) Polygon(ctx context.Context, name string) (*Polygon, error) {
	p, err := newPath(name, polytype)
	if err != nil {
		return &Polygon{}, err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return &Polygon{}, ErrDownloadFailed{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		}
	}

	if res.StatusCode() >= 400 {
		return &Polygon{}, ErrDownloadFailed{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		}
	}
	defer res.Close()

	polygon := NewPolygon(name, res.RawBody())
	err = polygon.Process()
	if err != nil {
		return &Polygon{}, err
	}

	return polygon, nil
}

// Download a dataset to output path
func (g *Geofabrik) Download(ctx context.Context, name string, outpath string) error {
	p, err := newPath(name, pbftype)
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf(
		"%s/%s",
		outpath,
		p.filename,
	)

	req := g.NR().SetHeader(
		"Accept",
		"application/octet-stream",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return ErrDownloadFailed{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		}
	}
	defer res.Close()

	if res.IsError() {
		return ErrDownloadFailed{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		}
	}

	err = g.writeOrRemove(filepath, res, func(w io.Writer) error {
		_, err := w.Write(res.Body())
		return err
	})
	if err != nil {
		return ErrCopyFailed{
			Message: err.Error(),
		}
	}

	return nil
}

func (g *Geofabrik) writeOrRemove(dest string, res *rip.Response, write func(w io.Writer) error) (err error) {
	tDir := tmpDir(dest)
	if _, err := os.Stat(tDir); os.IsNotExist(err) {
		defer func() {
			if err != nil {
				os.RemoveAll(tDir)
			}
		}()

		err = os.MkdirAll(tDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("while creating temporary directory: %w", err)
		}
	}

	f, err := os.CreateTemp(tDir, "tmp-")
	if err != nil {
		return fmt.Errorf("while creating temporary file: %w", err)
	}

	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	bufw := bufio.NewWriter(f)
	w := io.Writer(bufw)

	if g.withProgress {
		g.progress.reset()
		g.progress.setTotalByte(res.ContentLength())

		w = io.MultiWriter(bufw, g.progress)
	}

	if err := write(w); err != nil {
		return fmt.Errorf("while writing to temporary file: %w", err)
	}

	if err := bufw.Flush(); err != nil {
		return fmt.Errorf("while flushing bufwriter: %w", err)
	}

	if err := f.Chmod(0644); err != nil {
		return fmt.Errorf("while changing mode of file: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("while syncing content to storage: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("while closing temporary file: %w", err)
	}

	return os.Rename(f.Name(), dest)
}

func tmpDir(dest string) string {
	tmpDir := os.Getenv("GEOFABRIK_TMPDIR")
	if tmpDir == "" {
		tmpDir = filepath.Dir(dest)
	}

	return tmpDir
}
