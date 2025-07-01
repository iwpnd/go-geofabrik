package geofabrik

import (
	"fmt"
	"strings"
)

type Path struct {
	name     string
	uri      string
	filename string
}

func newPath(name string, ftype FileType) (*Path, error) {
	p := &Path{name: name}

	if err := p.process(ftype); err != nil {
		return &Path{}, err
	}

	return p, nil
}

func (p *Path) validate() error {
	if p.name == "" {
		return &EmptyNameError{}
	}

	return nil
}

func (p *Path) process(ftype FileType) error {
	if err := p.validate(); err != nil {
		return err
	}

	// sanitize start and end
	p.name = strings.TrimPrefix(p.name, "/")
	p.name = strings.TrimSuffix(p.name, "/")

	elements := strings.Split(p.name, "/")
	if len(elements) == 1 {
		switch ftype { //nolint: exhaustive
		case polytype:
			p.filename = fmt.Sprintf("%s%s", elements[0], ftype)
		default:
			p.filename = fmt.Sprintf("%s-latest%s", elements[0], ftype)
		}
		p.uri = fmt.Sprintf("/%s", p.filename)
		return nil
	}

	switch ftype { //nolint: exhaustive
	case polytype:
		p.filename = fmt.Sprintf(
			"%s%s",
			elements[len(elements)-1],
			ftype,
		)
	default:
		p.filename = fmt.Sprintf(
			"%s-latest%s",
			elements[len(elements)-1],
			ftype,
		)
	}

	p.uri = fmt.Sprintf(
		"/%s/%s",
		strings.Join(
			elements[0:len(elements)-1], "/",
		),
		p.filename,
	)

	return nil
}
