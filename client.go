package geofabrik

import (
	"fmt"
	"io"
	"os"
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
func New(host string, withProgress bool, options ...rip.Option) (*Geofabrik, error) {
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
func (g *Geofabrik) MD5(name string) (string, error) {
	p, err := newPath(name, md5type)
	if err != nil {
		return "", err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
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

func (g *Geofabrik) Polygon(name string) (*Polygon, error) {
	p, err := newPath(name, polytype)
	if err != nil {
		return &Polygon{}, err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
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
		// TODO: add error to errors.go
		return &Polygon{}, err
	}

	return polygon, nil
}

// Download a dataset to output path
func (g *Geofabrik) Download(name string, outpath string) error {
	p, err := newPath(name, pbftype)
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf(
		"%s/%s",
		outpath,
		p.filename,
	)

	// TODO: sanitize outpath
	out, err := os.Create(filepath)
	if err != nil {
		return ErrCreateFile{Message: err.Error()}
	}
	defer out.Close()

	req := g.NR().SetHeader(
		"Accept",
		"application/octet-stream",
	)
	res, err := req.Execute(
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

	if g.withProgress {
		g.progress.reset()
		g.progress.setTotalByte(res.ContentLength())

		mr := io.MultiWriter(out, g.progress)
		_, err := io.Copy(mr, res.RawBody())
		if err != nil {
			return ErrCreateFile{
				Message: err.Error(),
			}
		}
		return nil
	}

	_, err = io.Copy(out, res.RawBody())
	if err != nil {
		if err != nil {
			return ErrCopyFailed{
				Message: err.Error(),
			}
		}
	}

	return nil
}
