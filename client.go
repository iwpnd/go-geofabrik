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
	md5type FileType = ".osm.pbf.md5"
	pbftype FileType = ".osm.pbf"
)

type Geofabrik struct {
	*rip.Client
	withProgress bool
	progress     *Progress
}

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

func (g *Geofabrik) LatestMD5(name string) (string, error) {
	p, err := newPath(name)
	if err != nil {
		return "", err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
		"GET",
		fmt.Sprintf("%s%s", p.uri, md5type),
	)
	if err != nil || !res.IsSuccess() {
		return "", &ErrDownloadFailed{
			msg:  err.Error(),
			code: res.StatusCode(),
		}
	}
	defer res.Close()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}

func (g *Geofabrik) Download(name string, outpath string) error {
	p, err := newPath(name)
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf(
		"%s/%s%s",
		outpath,
		p.filename,
		pbftype,
	)

	// TODO: sanitize outpath
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf(
			"could not create out file %s: %s",
			filepath,
			err.Error(),
		)
	}
	defer out.Close()

	req := g.NR().SetHeader(
		"Accept",
		"application/octet-stream",
	)
	res, err := req.Execute(
		"GET",
		fmt.Sprintf("%s%s", p.uri, pbftype),
	)
	if err != nil || !res.IsSuccess() {
		return &ErrDownloadFailed{
			msg:  err.Error(),
			code: res.StatusCode(),
		}
	}
	defer res.Close()

	if g.withProgress {
		g.progress.reset()
		g.progress.setTotalByte(res.ContentLength())

		mr := io.MultiWriter(out, g.progress)
		_, err := io.Copy(mr, res.RawResponse.Body)
		if err != nil {
			return &ErrCreateFile{
				msg: err.Error(),
			}
		}
		return nil
	}

	_, err = io.Copy(out, res.RawBody())
	if err != nil {
		if err != nil {
			return &ErrCopyFailed{
				msg: err.Error(),
			}
		}
	}

	return nil
}
