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

func (tl TrackList) String() string {
	return tl.StringWithOffset(0)
}

func (tl TrackList) StringWithOffset(offset int) string {
	var sb strings.Builder

	sb.WriteString("```\n")
	for i, s := range tl {
		fmt.Fprintf(&sb, "%d) %s\n", i+1+offset, s)
	}
	sb.WriteString("```")
	return sb.String()
}
