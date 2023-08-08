package geofabrik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizePath(t *testing.T) {
	type tcase struct {
		input            string
		ftype            FileType
		expectedUri      string
		expectedFileName string
	}

	tests := map[string]tcase{
		"should tokenize one level": {
			input:            "europe",
			ftype:            pbftype,
			expectedUri:      "/europe-latest.osm.pbf",
			expectedFileName: "europe-latest.osm.pbf",
		},
		"should tokenize two levels": {
			input:            "europe/germany",
			ftype:            pbftype,
			expectedUri:      "/europe/germany-latest.osm.pbf",
			expectedFileName: "germany-latest.osm.pbf",
		},
		"should tokenize three levels": {
			input:            "europe/germany/berlin",
			ftype:            pbftype,
			expectedUri:      "/europe/germany/berlin-latest.osm.pbf",
			expectedFileName: "berlin-latest.osm.pbf",
		},
		"should persist other seperators": {
			input:            "europe/ireland-and-northern-ireland",
			ftype:            pbftype,
			expectedUri:      "/europe/ireland-and-northern-ireland-latest.osm.pbf",
			expectedFileName: "ireland-and-northern-ireland-latest.osm.pbf",
		},
		"should sanitize input": {
			input:            "/europe/",
			ftype:            pbftype,
			expectedUri:      "/europe-latest.osm.pbf",
			expectedFileName: "europe-latest.osm.pbf",
		},
	}

	for _, test := range tests {
		p, err := newPath(test.input, test.ftype)
		if err != nil {
			t.Fatalf("failed to create valid path: %v", err.Error())
		}
		assert.Equal(t, test.expectedUri, p.uri)
		assert.Equal(t, test.expectedFileName, p.filename)
	}
}
