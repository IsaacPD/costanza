package youtube

import (
	"encoding/json"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/sound/sources/local"
)

const (
	WATCH_FORMAT = "http://youtube.com/watch?v=%s"
)

type (
	Format struct {
		Url      string  `json:"url"`
		FormatID string  `json:"format_id"`
		Vcodec   string  `json:"vcodec"`
		Acodec   string  `json:"acodec"`
		Abr      float32 `json:"abr"`
	}
	videoDetails struct {
		Formats  []Format `json:"formats"`
		Title    string   `json:"title"`
		Uploader string   `json:"uploader"`
		Duration int      `json:"duration"`
	}
)

func getDetails(id string) videoDetails {
	cmd := exec.Command("youtube-dl", "--skip-download", "--print-json", id)
	out, _ := cmd.Output()

	var resp videoDetails
	err := json.Unmarshal(out, &resp)
	if err != nil {
		logrus.Warnf("Error unmarshalling youtube-dl json output for %s", id)
	}
	return resp
}

func getBestAudioFormat(formats []Format) Format {
	var best Format
	for _, f := range formats {
		if f.Vcodec != "none" {
			continue
		}
		if f.Abr > best.Abr {
			best = f
		}
	}
	return best
}

func cmd(id, formatID string) (dl *exec.Cmd, ffmpeg *exec.Cmd) {
	dl = exec.Command("youtube-dl", "-f", formatID, "-o", "-", id)
	ffmpeg = local.Ffmpeg("-")

	r, w := io.Pipe()

	dl.Stdout = w
	ffmpeg.Stdin = r
	return
}
