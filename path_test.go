package geofabrik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizePath(t *testing.T) {
	type tcase struct {
		input            string
		expectedUri      string
		expectedFileName string
	}

	tests := map[string]tcase{
		"should tokenize one level": {
			input:            "europe",
			expectedUri:      "/europe-latest",
			expectedFileName: "europe-latest",
		},
		"should tokenize two levels": {
			input:            "europe/germany",
			expectedUri:      "/europe/germany-latest",
			expectedFileName: "germany-latest",
		},
		"should tokenize three levels": {
			input:            "europe/germany/berlin",
			expectedUri:      "/europe/germany/berlin-latest",
			expectedFileName: "berlin-latest",
		},
		"should persist other seperators": {
			input:            "europe/ireland-and-northern-ireland",
			expectedUri:      "/europe/ireland-and-northern-ireland-latest",
			expectedFileName: "ireland-and-northern-ireland-latest",
		},
		"should sanitize input": {
			input:            "/europe/",
			expectedUri:      "/europe-latest",
			expectedFileName: "europe-latest",
		},
	}

	for _, test := range tests {
		p, err := newPath(test.input)
		if err != nil {
			t.Fatalf("failed to create valid path: %v", err.Error())
		}
		assert.Equal(t, test.expectedUri, p.uri)
		assert.Equal(t, test.expectedFileName, p.filename)
	}
}
