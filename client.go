package client

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

type ErrMissingName struct {
}

func (e *ErrMissingName) Error() string {
	return "missing dataset name"
}

type Geofabrik struct {
	*rip.Client
}

func tokenizePath(elements []string) string {
	if len(elements) == 1 {
		return fmt.Sprintf("/%s-latest", elements[0])
	}

	file := elements[len(elements)-1]
	path := strings.Join(elements[0:len(elements)-1], "/")

	return fmt.Sprintf("/%s/%s-latest", path, file)
}

func New(host string, options ...rip.Option) (*Geofabrik, error) {
	c, err := rip.NewClient(host, options...)
	if err != nil {
		return &Geofabrik{}, err
	}
	return &Geofabrik{c}, nil
}

func (g *Geofabrik) LatestMD5(name string) (string, error) {
	elements := strings.Split(name, "/")
	path := tokenizePath(elements)
	req := g.NR().SetHeader("Accept", "text/plain; charset=utf-8")
	if path == "" {
		return "", &ErrMissingName{}
	}

	res, err := req.Execute("GET", fmt.Sprintf("%s.%s", path, md5type))
	if err != nil {
		return "", err
	}
	defer res.Close()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}

func (g *Geofabrik) SimpleDownload(name string, outpath string) error {
	elements := strings.Split(name, "/")
	path := tokenizePath(elements)
	if path == "" {
		return &ErrMissingName{}
	}

	file := elements[len(elements)-1]
	filepath := fmt.Sprintf("%s/%s-latest%s", outpath, file, pbftype)
	// TODO: sanitize outpath
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("could not create out file %s%s%s: %s", outpath, path, pbftype, err.Error())
	}
	defer out.Close()

	req := g.NR().SetHeader("Accept", "application/octet-stream")
	res, err := req.Execute("GET", fmt.Sprintf("%s%s", path, pbftype))
	if err != nil {
		return err
	}
	defer res.Close()

	if !res.IsSuccess() {
		return fmt.Errorf("download unsuccessful: %v", res.StatusCode())
	}

	_, err = io.Copy(out, res.RawBody())
	if err != nil {
		fmt.Println("failed: ", err.Error())
		return err
	}

	return nil
}
