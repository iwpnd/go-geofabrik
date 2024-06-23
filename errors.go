package geofabrik

import "fmt"

type ErrEmptyName struct{}

func (e ErrEmptyName) Error() string {
	return "name is empty"
}

type ErrDownloadFailed struct {
	URL     string
	Message string
	Code    int
}

func (e ErrDownloadFailed) Error() string {
	return fmt.Sprintf(
		"download failed with: %s (code: %d)",
		e.Message,
		e.Code,
	)
}

type ErrCopyFailed struct {
	Message string
}

func (e ErrCopyFailed) Error() string {
	return fmt.Sprintf("failed to save file: %s", e.Message)
}
