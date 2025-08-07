package geofabrik

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iwpnd/rip"
)

type FileType string

const (
	md5type  FileType = ".osm.pbf.md5"
	pbftype  FileType = ".osm.pbf"
	polytype FileType = ".poly"
)

// Geofabrik wraps a rest client.
type Geofabrik struct {
	*rip.Client
}

// New is the constructor for a Geofabrik.
func New(host string, options ...rip.Option) (*Geofabrik, error) {
	c, err := rip.NewClient(host, options...)
	if err != nil {
		return &Geofabrik{}, err
	}

	return &Geofabrik{
		Client: c,
	}, nil
}

// MD5 will return the latest MD5 of a dataset
func (g *Geofabrik) MD5(ctx context.Context, name string) (string, error) {
	p, err := newPath(name, md5type)
	if err != nil {
		return "", err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return "", errors.Join(err, DownloadFailedError{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		})
	}

	if res.StatusCode() >= 400 {
		return "", errors.Join(err, DownloadFailedError{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		})
	}
	defer func() {
		if cErr := res.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				err = errors.Join(err, cErr)
			}
		}
	}()

	md5 := strings.Split(res.String(), "  ")[0]

	return md5, nil
}

func (g *Geofabrik) Polygon(ctx context.Context, name string) (*Polygon, error) {
	p, err := newPath(name, polytype)
	if err != nil {
		return &Polygon{}, err
	}

	req := g.NR().SetHeader(
		"Accept",
		"text/plain; charset=utf-8",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return &Polygon{}, errors.Join(err, DownloadFailedError{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		})
	}

	if res.StatusCode() >= 400 {
		return &Polygon{}, DownloadFailedError{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		}
	}
	defer func() {
		if cErr := res.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				err = errors.Join(err, cErr)
			}
		}
	}()

	polygon := NewPolygon(name, res.RawBody())
	err = polygon.Process()
	if err != nil {
		return &Polygon{}, err
	}

	return polygon, nil
}

// Download a dataset to output path
func (g *Geofabrik) Download(ctx context.Context, name, outpath string) error {
	p, err := newPath(name, pbftype)
	if err != nil {
		return err
	}

	fp := fmt.Sprintf(
		"%s/%s",
		outpath,
		p.filename,
	)

	req := g.NR().SetHeader(
		"Accept",
		"application/octet-stream",
	)
	res, err := req.Execute(
		ctx,
		"GET",
		p.uri,
	)
	if err != nil {
		return errors.Join(err, DownloadFailedError{
			Message: err.Error(),
			Code:    res.StatusCode(),
			URL:     res.Request.URL,
		})
	}
	defer func() {
		if cErr := res.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				err = errors.Join(err, cErr)
			}
		}
	}()

	if res.IsError() {
		return DownloadFailedError{
			Code: res.StatusCode(),
			URL:  res.Request.URL,
		}
	}

	err = g.writeOrRemove(ctx, fp, func(w io.Writer) error {
		_, err := io.Copy(w, res.RawBody())
		return err
	})
	if err != nil {
		return errors.Join(err, CopyFailedError{
			Message: err.Error(),
		})
	}

	return nil
}

func (g *Geofabrik) writeOrRemove(ctx context.Context, dest string, write func(w io.Writer) error) (err error) {
	tDir := tmpDir(dest)
	if _, statErr := os.Stat(tDir); os.IsNotExist(statErr) {
		if mkErr := os.MkdirAll(tDir, 0o750); mkErr != nil {
			return fmt.Errorf("creating temporary directory %q: %w", tDir, mkErr)
		}
	}

	f, err := os.CreateTemp(tDir, "tmp-")
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}

	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
			_ = os.RemoveAll(tDir)
		}
	}()

	// instead of writing directly to f, we set up a pipe:
	pr, pw := io.Pipe()
	defer pr.Close()

	// writer goroutine
	done := make(chan error, 1)
	go func() {
		bufw := bufio.NewWriter(pw)
		w := io.Writer(bufw)

		if err := write(w); err != nil {
			pw.CloseWithError(fmt.Errorf("while writing to pipe: %w", err))
			done <- err
			return
		}
		if err := bufw.Flush(); err != nil {
			pw.CloseWithError(fmt.Errorf("while flushing: %w", err))
			done <- err
			return
		}
		pw.Close()
		done <- nil
	}()

	// copy from the pipe into our temp file, and watch ctx.Done()
	copyErrCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(f, pr)
		copyErrCh <- err
	}()

	select {
	case <-ctx.Done():
		// user cancelled: clean up
		pw.CloseWithError(ctx.Err())
		f.Close()
		os.Remove(f.Name())
		return ctx.Err()
	case err := <-done:
		if err != nil {
			// writer goroutine failed early
			f.Close()
			os.Remove(f.Name())
			return err
		}
		// writer finished; wait for the final copy into f
		if err := <-copyErrCh; err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}
	}

	// sync and rename
	if err = f.Sync(); err != nil {
		return fmt.Errorf("while syncing content to storage: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("while closing temporary file: %w", err)
	}
	return os.Rename(f.Name(), dest)
}

func tmpDir(dest string) string {
	tmpDir := os.Getenv("GEOFABRIK_TMPDIR")
	if tmpDir == "" {
		tmpDir = filepath.Dir(dest)
	}

	return tmpDir
}
