package geofabrik

import (
	"fmt"
	"strings"
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

	// sanitize start and end
	p.name = strings.TrimPrefix(p.name, "/")
	p.name = strings.TrimSuffix(p.name, "/")

	elements := strings.Split(p.name, "/")
	if len(elements) == 1 {
		p.filename = fmt.Sprintf("%s-latest", elements[0])
		p.uri = fmt.Sprintf("/%s", p.filename)
		return nil
	}

	p.filename = fmt.Sprintf(
		"%s-latest",
		elements[len(elements)-1],
	)
	p.uri = fmt.Sprintf(
		"/%s/%s",
		strings.Join(
			elements[0:len(elements)-1], "/",
		),
		p.filename,
	)

	return nil
}
