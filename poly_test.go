package geofabrik

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolygonProcess(t *testing.T) {
	type tcase struct {
		name     string
		input    []byte
		expected string
	}

	tests := map[string]tcase{
		"polygon": {
			name:     "TestPolygon",
			input:    []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"),
			expected: `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{"name":"test"}}`,
		},
		"multipolygon": {
			name:     "TestMultiPolygon",
			input:    []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\ntest2\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"),
			expected: `{"type":"Feature","geometry":{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,1],[0,0]]],[[[0,0],[1,0],[1,1],[0,1],[0,0]]]]},"properties":{"name":"test"}}`,
		},
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			r := bytes.NewReader(tc.input)
			p := NewPolygon(tc.name, r)
			err := p.Process()
			if err != nil {
				t.Fatal("could not process polygon", err)
			}

			properties := map[string]interface{}{"name": "test"}
			f, err := p.ToFeature(properties)
			if err != nil {
				t.Fatal("failed to build feature", err)
			}

			assert.Equal(t, tc.expected, f)
		}
	}

	for _, test := range tests {
		fn(test)
	}
}
