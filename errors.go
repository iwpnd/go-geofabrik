package geofabrik

import "fmt"

type EmptyNameError struct{}

func (e EmptyNameError) Error() string {
	return "name is empty"
}

type DownloadFailedError struct {
	URL     string
	Message string
	Code    int
}

func (e DownloadFailedError) Error() string {
	return fmt.Sprintf(
		"download failed with: %s (code: %d)",
		e.Message,
		e.Code,
	)
}

type CopyFailedError struct {
	Message string
}

func (e CopyFailedError) Error() string {
	return fmt.Sprintf("failed to save file: %s", e.Message)
}
