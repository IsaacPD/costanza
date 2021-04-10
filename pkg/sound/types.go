package sound

import (
	"fmt"
	"io"
	"strings"
)

type Track interface {
	GetReader() (io.Reader, error)
	Start() error
	Stop()
}

type TrackList []Track

func (tl TrackList) String(offset int) string {
	var sb strings.Builder

	sb.WriteString("```nim\n")
	for i, s := range tl {
		fmt.Fprintf(&sb, "%d) %s\n", i+1+offset, s)
	}
	sb.WriteString("```")
	return sb.String()
}
