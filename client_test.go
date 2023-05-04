package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ts *httptest.Server
)

func setupTestServer() func() {
	ts = httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				switch r.URL.Path {
				case "/foo-latest.osm.pbf.md5":
					{
						accept := r.Header.Get("Accept")
						switch accept {
						case "text/plain; charset=utf-8":
							w.WriteHeader(http.StatusOK)
							w.Header().Set("Content-Type", "text/plain; charset=utf-8")
							fmt.Fprint(w, "bar  foo")
						default:
							w.WriteHeader(http.StatusNotAcceptable)
							fmt.Fprint(w, "nope")
						}
					}
				default:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
				}
			}
		}))

	return func() {
		ts.Close()
	}
}

func TestTokenizePath(t *testing.T) {
	type tcase struct {
		name     string
		expected string
	}

	tests := map[string]tcase{
		"should tokenize one level": {
			name:     "europe",
			expected: "/europe-latest",
		},
		"should tokenize two levels": {
			name:     "europe/germany",
			expected: "/europe/germany-latest",
		},
		"should tokenize three levels": {
			name:     "europe/germany/berlin",
			expected: "/europe/germany/berlin-latest",
		},
		"should persist other seperators": {
			name:     "europe/ireland-and-northern-ireland",
			expected: "/europe/ireland-and-northern-ireland-latest",
		},
	}

	for _, test := range tests {
		got := tokenizePath(test.name)
		assert.Equal(t, test.expected, got)
	}
}

func TestGetMD5(t *testing.T) {
	teardown := setupTestServer()
	defer teardown()

	g, err := New(ts.URL)

	if err != nil {
		t.Error("could not initialize client")
	}

	type tcase struct {
		name     string
		expected string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			got, err := g.LatestMD5("foo")
			if err != nil {
				t.Fatal("failed to get md5")
			}

			assert.Equal(t, tc.expected, got)
		}
	}

	tests := map[string]tcase{
		"should resolve md5": {
			name:     "foo",
			expected: "bar",
		},
		"should fail to resolve md5": {
			name:     "test",
			expected: "",
		},
	}

	for _, test := range tests {
		fn(test)
	}
}
