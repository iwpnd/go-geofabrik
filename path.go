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
		ds := elements[0]
		switch ftype { //nolint: exhaustive
		case polytype:
			p.uri = fmt.Sprintf("/%s%s", ds, ftype)
		default:
			p.uri = fmt.Sprintf("/%s-latest%s", ds, ftype)
		}
		p.filename = fmt.Sprintf("%s%s", ds, ftype)
		return nil
	}

	ds := elements[len(elements)-1]
	var f string
	switch ftype { //nolint: exhaustive
	case polytype:
		f = fmt.Sprintf(
			"%s%s",
			ds,
			ftype,
		)
	default:
		f = fmt.Sprintf(
			"%s-latest%s",
			ds,
			ftype,
		)
	}

	p.uri = fmt.Sprintf(
		"/%s/%s",
		strings.Join(
			elements[0:len(elements)-1], "/",
		),
		f,
	)

	p.filename = fmt.Sprintf(
		"%s%s",
		ds,
		ftype,
	)

	return nil
}
