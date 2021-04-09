package local

import (
	"io"
	"os/exec"
	"strconv"

	"github.com/isaacpd/costanza/pkg/util"
)

type localTrack struct {
	Path string

	cmd  *exec.Cmd
	next *exec.Cmd
}

func Ffmpeg(song string) *exec.Cmd {
	return exec.Command("ffmpeg", "-i", song, "-f", "s16le", "-ar", strconv.Itoa(util.FRAME_RATE), "-ac",
		strconv.Itoa(util.CHANNELS), "pipe:1")
}

func NewLocalTrack(path string) *localTrack {
	return &localTrack{
		Path: path,
		cmd:  Ffmpeg(path),
	}
}

func (lt *localTrack) GetReader() (io.Reader, error) {
	if lt.next != nil {
		lt.cmd = lt.next
	}
	lt.next = Ffmpeg(lt.Path)
	return lt.cmd.StdoutPipe()
}

func (lt *localTrack) Start() error {
	return lt.cmd.Start()
}

func (lt *localTrack) Stop() {
	_ = lt.cmd.Process.Kill()
}

func (lt *localTrack) String() string {
	return lt.Path
}
