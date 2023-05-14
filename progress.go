package geofabrik

import "fmt"

type Progress struct {
	receivedBytes int64
	totalBytes    int64
}

func NewProgress(totalBytes int64) *Progress {
	return &Progress{
		totalBytes: totalBytes,
	}
}

func (p *Progress) reset() {
	p.totalBytes = 0
	p.receivedBytes = 0
}

func (p *Progress) setTotalByte(b int64) {
	p.totalBytes = b
}

func (p *Progress) Read(b []byte) (int, error) {
	i := len(b)
	p.progress(i)
	return i, nil
}

func (p *Progress) Write(b []byte) (int, error) {
	i := len(b)
	p.progress(i)
	return i, nil
}

func (p *Progress) progress(i int) {
	p.receivedBytes += int64(i)
	percent := float64(p.receivedBytes) / float64(p.totalBytes) * 100
	bar := "["
	for j := 0; j < 50; j++ {
		if float64(j)/50*100 <= percent {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	fmt.Printf("\r%s %.2f%% (%d/%d bytes)", bar, percent, p.receivedBytes, p.totalBytes)
}
