package youtube

import (
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

func NewYoutubeTrack(id string) (*youtubeTrack, error) {
	details, err := getDetails(id)
	if err != nil {
		return nil, err
	}

	format := getBestAudioFormat(details.Formats)
	dur, _ := time.ParseDuration(fmt.Sprintf("%ds", details.Duration))
	logrus.Tracef("Best format for %s is %+v", details.Title, format)
	return &youtubeTrack{
		ID:       id,
		Title:    details.Title,
		Uploader: details.Uploader,
		Length:   dur.String(),
		formatID: format.FormatID,
		cmd:      cmd(id, format.FormatID),
	}, nil
}

func (yt *youtubeTrack) GetReader() (io.Reader, error) {
	if yt.next != nil {
		yt.cmd = yt.next
	}
	yt.next = cmd(yt.ID, yt.formatID)
	return yt.cmd.StdoutPipe()
}

func (yt *youtubeTrack) Start() (err error) {
	return yt.cmd.Start()
}

func (yt *youtubeTrack) Stop() {
	yt.cmd.Process.Kill()
}
