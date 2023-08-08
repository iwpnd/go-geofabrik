package geofabrik

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolygonProcess(t *testing.T) {
	type tcase struct {
		name       string
		properties map[string]interface{}
		input      []byte
		expected   string
	}

	tests := map[string]tcase{
		"polygon": {
			name:     "TestPolygon",
			input:    []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"),
			expected: `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{"name":"TestPolygon"}}`,
		},
		"polygon with properties": {
			name:       "TestPolygon",
			properties: map[string]interface{}{"foo": "bar"},
			input:      []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"),
			expected:   `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{"foo":"bar"}}`,
		},
		"multipolygon": {
			name:     "TestMultiPolygon",
			input:    []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\ntest2\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"),
			expected: `{"type":"Feature","geometry":{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,1],[0,0]]],[[[0,0],[1,0],[1,1],[0,1],[0,0]]]]},"properties":{"name":"TestMultiPolygon"}}`,
		},
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			r := bytes.NewReader(tc.input)
			p := NewPolygon(tc.name, r)
			if tc.properties != nil {
				p.WithProperties(tc.properties)
			}
			err := p.Process()
			if err != nil {
				t.Fatal("could not process polygon", err)
			}

			f, err := p.ToFeature()
			if err != nil {
				t.Fatal("failed to build feature", err)
			}

			assert.Equal(t, tc.expected, f)
		}
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
