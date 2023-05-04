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

type ErrEmptyName struct{}

func (e *ErrEmptyName) Error() string {
	return "name is empty"
}

type Path struct {
	name     string
	uri      string
	filename string
}

func newPath(name string) (*Path, error) {
	p := &Path{name: name}

	if err := p.process(); err != nil {
		return &Path{}, err
	}

	return p, nil
}

func (p *Path) validate() error {
	if p.name == "" {
		return &ErrEmptyName{}
	}

	return nil
}

func (p *Path) process() error {
	if err := p.validate(); err != nil {
		return err
	}

	// sanitize stat and end
	p.name = strings.TrimPrefix(p.name, "/")
	p.name = strings.TrimSuffix(p.name, "/")

	elements := strings.Split(p.name, "/")
	if len(elements) == 1 {
		p.filename = fmt.Sprintf("%s-latest", elements[0])
		p.uri = fmt.Sprintf("/%s", p.filename)
		return nil
	}

	p.filename = fmt.Sprintf("%s-latest", elements[len(elements)-1])
	p.uri = fmt.Sprintf("/%s/%s", strings.Join(elements[0:len(elements)-1], "/"), p.filename)

	return nil
}

type Geofabrik struct {
	*rip.Client
}

func New(host string, options ...rip.Option) (*Geofabrik, error) {
	c, err := rip.NewClient(host, options...)
	if err != nil {
		return &Geofabrik{}, err
	}
	return &Geofabrik{c}, nil
}

func (g *Geofabrik) LatestMD5(name string) (string, error) {
	p, err := newPath(name)
	if err != nil {
		return "", err
	}

	req := g.NR().SetHeader("Accept", "text/plain; charset=utf-8")
	res, err := req.Execute("GET", fmt.Sprintf("%s.%s", p.uri, md5type))
	if err != nil {
		return "", err
	}
	defer res.Close()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}

func (g *Geofabrik) SimpleDownload(name string, outpath string) error {
	p, err := newPath(name)
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/%s%s", outpath, p.filename, pbftype)
	// TODO: sanitize outpath
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("could not create out file %s: %s", filepath, err.Error())
	}
	defer out.Close()

	req := g.NR().SetHeader("Accept", "application/octet-stream")
	res, err := req.Execute("GET", fmt.Sprintf("%s%s", p.uri, pbftype))
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
