package geofabrik

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ts *httptest.Server
)

func fileExists(dir string, filename string) bool {
	info, err := os.Stat(fmt.Sprintf("%s/%s", dir, filename))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func randomDataOfSize(size int) []byte {
	return []byte(strings.Repeat("#", size))
}

func compareHash(t *testing.T, expected, got []byte) bool {
	expectedHash := md5.Sum(expected)
	gotHash := md5.Sum(got)

	if expectedHash != gotHash {
		t.Errorf("expected: %s, got: %s", expectedHash, gotHash)
	}

	return expectedHash == gotHash
}

func setupTestServer(responseData []byte) func() {
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
				case "/foo-latest.osm.pbf":
					{
						accept := r.Header.Get("Accept")
						switch accept {
						case "application/octet-stream":
							w.WriteHeader(http.StatusOK)
							w.Header().Set("Content-Type", "application/octet-stream")
							if responseData != nil {
								reader := bytes.NewReader(responseData)
								_, err := io.Copy(w, reader)
								if err != nil {
									fmt.Println("failed to send response: ", err.Error())
									return
								}
								return
							}

							reader := bytes.NewReader(randomDataOfSize(1028 * 128))
							_, err := io.Copy(w, reader)
							if err != nil {
								fmt.Println("failed to send response: ", err.Error())
								return
							}
							return
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

func TestGetMD5(t *testing.T) {
	teardown := setupTestServer(nil)
	defer teardown()

	g, err := New(ts.URL, false)
	if err != nil {
		t.Fatal("could not initialize client")
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

func TestSimpleDownload(t *testing.T) {
	dir, err := os.MkdirTemp(".", "tmp")
	if err != nil {
		t.Fatalf("error creating temp directory: %s", err)
	}
	defer os.RemoveAll(dir)

	responseFile := randomDataOfSize(1024 * 128)
	teardown := setupTestServer(responseFile)
	defer teardown()

	g, err := New(ts.URL, false)
	if err != nil {
		t.Fatal("could not initialize client")
	}

	err = g.SimpleDownload("foo", dir)
	if err != nil {
		t.Fatalf("failed to download: %v", err.Error())
	}

	testfile := "foo-latest.osm.pbf"
	assert.Equal(t, true, fileExists(dir, testfile))

	got, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, testfile))
	if err != nil {
		t.Fatalf("could not open test file: %s/%s", dir, testfile)
	}
	assert.Equal(t, true, compareHash(t, responseFile, got))
}
