package youtube

import (
	"io"

	"github.com/sirupsen/logrus"
)

func NewYoutubeTrack(id string) *youtubeTrack {
	details := getDetails(id)
	format := getBestAudioFormat(details.Formats)
	logrus.Tracef("Best format for %s is %+v", details.Title, format)
	dl, ffmpeg := cmd(id, format.FormatID)
	return &youtubeTrack{
		ID:     id,
		Title:  details.Title,
		ffmpeg: ffmpeg,
		dl:     dl,
	}
}

func (yt *youtubeTrack) GetReader() (io.Reader, error) {
	return yt.ffmpeg.StdoutPipe()
}

func (yt *youtubeTrack) Start() (err error) {
	err = yt.dl.Start()
	if err != nil {
		return
	}
	return yt.ffmpeg.Start()
}

func (yt *youtubeTrack) Stop() {
	yt.dl.Process.Kill()
	yt.ffmpeg.Process.Kill()
}
