package geofabrik

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iwpnd/rip"
	"github.com/stretchr/testify/assert"
)

var ts *httptest.Server

func fileExists(dir, filename string) bool {
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
	t.Helper()
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
			if r.Method == http.MethodGet {
				switch r.URL.Path {
				case "/foo-latest.osm.pbf.md5":
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
				case "/foo.poly":
					accept := r.Header.Get("Accept")
					switch accept {
					case "text/plain; charset=utf-8":
						w.WriteHeader(http.StatusOK)
						w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						fmt.Fprint(w, "test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND") //nolint: dupword
					default:
						w.WriteHeader(http.StatusNotAcceptable)
						fmt.Fprint(w, "nope")
					}
				case "/foo-latest.osm.pbf":
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
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}
		}))

	return func() {
		ts.Close()
	}
}

func TestGetPolygon(t *testing.T) {
	teardown := setupTestServer(nil)
	defer teardown()

	g, err := New(ts.URL)
	if err != nil {
		t.Fatal("could not initialize client")
	}

	type tcase struct {
		name     string
		input    []byte
		expected string
	}

	tests := map[string]tcase{
		"polygon": {
			name:     "foo",
			input:    []byte("test\ntest\n   0   0 \n   1   0\n   1   1\n   0   1\n   0   0\nEND\nEND"), //nolint: dupword
			expected: `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{"name":"foo"}}`,
		},
	}

	ctx := t.Context()

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			p, err := g.Polygon(ctx, tc.name)
			if err != nil {
				t.Fatal("failed to get polygon", err)
			}
			got, err := p.ToFeature()
			if err != nil {
				t.Fatal("failed to build feature", err)
			}

			assert.Equal(t, tc.expected, got)
		}
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestGetMD5(t *testing.T) {
	teardown := setupTestServer(nil)
	defer teardown()

	g, err := New(ts.URL)
	if err != nil {
		t.Fatal("could not initialize client")
	}

	ctx := t.Context()

	type tcase struct {
		name     string
		expected string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			got, err := g.MD5(ctx, tc.name)
			if err != nil && tc.expected != "" {
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

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestDownloadCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/europe/germany-latest.osm.pbf" {
			time.Sleep(time.Second)
			w.Header().Add("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OSM DATA"))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	g, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = g.Download(ctx, "europe/germany", ".")
	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got: %v", err)
	}
}

func TestDownload(t *testing.T) {
	dir := t.TempDir()

	responseFile := randomDataOfSize(1024 * 128)
	teardown := setupTestServer(responseFile)
	defer teardown()

	ctx := t.Context()

	g, err := New(ts.URL)
	if err != nil {
		t.Fatal("could not initialize client")
	}

	err = g.Download(ctx, "foo", dir)
	if err != nil {
		t.Fatal(err.Error())
	}

	testfile := "foo.osm.pbf"
	assert.True(t, fileExists(dir, testfile))

	got, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, testfile))
	if err != nil {
		t.Fatalf("could not open test file: %s/%s", dir, testfile)
	}
	assert.True(t, compareHash(t, responseFile, got))
}

func TestDownloadFailed(t *testing.T) {
	teardown := setupTestServer(nil)
	defer teardown()

	g, err := New(ts.URL)
	if err != nil {
		t.Fatal("could not initialize client")
	}
	dir := t.TempDir()
	ctx := t.Context()

	err = g.Download(ctx, "bar", dir)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != nil {
		var got DownloadFailedError
		isErrDownloadFailed := errors.As(err, &got)
		want := DownloadFailedError{URL: ts.URL + "/bar-latest.osm.pbf", Code: http.StatusNotFound}

		assert.True(t, isErrDownloadFailed)
		assert.Equal(t, want, got)
	}
}

func TestWriteOrRemove(t *testing.T) {
	g, err := New(ts.URL)
	if err != nil {
		t.Fatal("could not initialize client")
	}
	dir := t.TempDir()

	testfile := "foo.osm.pbf"
	err = g.writeOrRemove(testfile, &rip.Response{}, func(w io.Writer) error {
		return fmt.Errorf("something went wrong")
	})
	if err == nil {
		t.Fatal("expected ErrCopyFailed but got nil")
	}
	assert.False(t, fileExists(dir, testfile))
}
