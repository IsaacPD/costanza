package local

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/util"
)

type localTrack struct {
	Path string
	Name string

	cmd  *exec.Cmd
	next *exec.Cmd
}

func Ffmpeg(song string) *exec.Cmd {
	return exec.Command("ffmpeg", "-i", song, "-f", "s16le", "-ar", strconv.Itoa(util.FRAME_RATE), "-ac",
		strconv.Itoa(util.CHANNELS), "pipe:1")
}

func NewLocalTrack(path string) sound.TrackList {
	f, e := os.Stat(path)
	if e != nil {
		logrus.Warnf("Couldn't find file at path %s", path)
		return nil
	}
	var tl sound.TrackList
	if !f.IsDir() {
		tl = append(tl, createLocalTrack(f.Name(), path))
		return tl
	}
	names := getDirChildrenNames(path)
	logrus.Tracef("Got names %s", names)
	for _, n := range names {
		tl = append(tl, createLocalTrack(fmt.Sprintf("%s/%s", path, n), n))
	}
	return tl
}

func createLocalTrack(path, name string) *localTrack {
	return &localTrack{
		Path: path,
		Name: name,
		cmd:  Ffmpeg(path),
	}
}

func getDirChildrenNames(path string) []string {
	logrus.Tracef("Getting children of %s", path)
	var names []string
	info, _ := ioutil.ReadDir(path)
	for _, i := range info {
		if i.IsDir() {
			continue
		}
		names = append(names, i.Name())
	}
	return names
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
	return lt.Name
}
