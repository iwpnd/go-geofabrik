package geofabrik

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Polygon ..
type Polygon struct {
	Name       string
	rings      [][][]float64
	scanner    bufio.Scanner
	properties map[string]interface{}
}

// NewPolygon ..
func NewPolygon(name string, data io.Reader) *Polygon {
	return &Polygon{
		Name:       name,
		properties: map[string]interface{}{"name": name},
		rings:      [][][]float64{},
		scanner:    *bufio.NewScanner(data),
	}
}

// WithProperties attaches properties to the Polygon, defaults to {"name":p.name}
func (p *Polygon) WithProperties(properties map[string]interface{}) *Polygon {
	p.properties = properties
	return p
}

// Process iterates over the returned geofabrik .poly document and
// collects coordinate pairs
func (p *Polygon) Process() error {
	recording := false
	idx := 0

	for p.scanner.Scan() {
		line := p.scanner.Text()
		t := strings.Split(strings.TrimSpace(line), "   ")

		if line == "END" {
			if !recording {
				// end of file
				continue
			}

			recording = false
			idx = idx + 1
		} else if len(t) > 1 {
			recording = true
			coords, err := parseStringSlice(t)
			if err != nil {
				return err
			}
			err = p.addCoordinates(coords, idx)
			if err != nil {
				return err
			}
		} else {
			if recording {
				// already recording
				continue
			}
			recording = true
			p.rings = append(p.rings, [][]float64{})
		}
	}
	return nil
}

func (p *Polygon) addCoordinates(coords []float64, idx int) error {
	p.rings[idx] = append(p.rings[idx], coords)

	return nil
}

// ToFeature returns a feature string if p.rings is populated
func (p *Polygon) ToFeature() (string, error) {
	if len(p.rings) == 0 {
		return "", errors.New("no polygons to create feature from")
	}

	isMultiPolygon := len(p.rings) > 1

	feature := `{"type":"Feature","geometry":{"type":`
	if isMultiPolygon {
		multi := [][][][]float64{}
		for i, x := range p.rings {
			multi = append(multi, [][][]float64{})
			multi[i] = [][][]float64{x}
		}
		rs, _ := json.Marshal(multi)
		feature = feature + `"MultiPolygon","coordinates":` + string(rs) + `},`
	} else {
		r, _ := json.Marshal(p.rings)
		feature = feature + `"Polygon","coordinates":` + string(r) + `},`
	}

	data, _ := json.Marshal(p.properties)
	feature = feature + `"properties":` + string(data)
	feature = feature + `}`

	return feature, nil
}

func parseStringSlice(line []string) ([]float64, error) {
	var out []float64

	for _, s := range line {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return []float64{}, err
		}
		out = append(out, f)
	}
	return out, nil
}
