package geofabrik

import "fmt"

type ErrEmptyName struct{}

func (e *ErrEmptyName) Error() string {
	return "name is empty"
}

type ErrDownloadFailed struct {
	msg  string
	code int
}

func (e *ErrDownloadFailed) Error() string {
	return fmt.Sprintf(
		"download failed with: %s (code: %d)",
		e.msg,
		e.code,
	)
}

type ErrCreateFile struct {
	msg string
}

func (e *ErrCreateFile) Error() string {
	return fmt.Sprintf("failed to create file: %s", e.msg)
}

type ErrCopyFailed struct {
	msg string
}

func (e *ErrCopyFailed) Error() string {
	return fmt.Sprintf("failed to save file: %s", e.msg)
}
