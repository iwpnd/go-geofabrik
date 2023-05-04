package client

import (
	"fmt"
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

func tokenizePath(name string) string {
	split := strings.Split(name, "/")

	if len(split) == 1 {
		return fmt.Sprintf("/%s-latest", split[0])
	}

	file := split[len(split)-1]
	path := strings.Join(split[0:len(split)-1], "/")

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
	path := tokenizePath(name)
	req := g.NR().SetHeader("Accept", "text/plain; charset=utf-8")
	if path == "" {
		return "", &ErrMissingName{}
	}

	res, err := req.Execute("GET", fmt.Sprintf("%s.%s", path, md5type))
	if err != nil {
		return "", err
	}
	defer res.RawBody().Close()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}
