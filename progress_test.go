package geofabrik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	p := newProgress()
	p.setTotalByte(1000)

	assert.Equal(t, int64(1000), p.totalBytes)

	p.reset()

	assert.Equal(t, int64(0), p.totalBytes)
	assert.Equal(t, int64(0), p.receivedBytes)

	_, err := p.Read(randomDataOfSize(1000))
	if err != nil {
		t.Fatal("could not read")
	}
	assert.Equal(t, int64(1000), p.receivedBytes)

	p.reset()

	_, err = p.Write(randomDataOfSize(1000))
	if err != nil {
		t.Fatal("could not read")
	}
	assert.Equal(t, int64(1000), p.receivedBytes)
}
