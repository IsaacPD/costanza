package youtube

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
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

func getDetails(id string) (videoDetails, error) {
	cmd := exec.Command("youtube-dl", "--skip-download", "--print-json", id)
	out, err := cmd.Output()
	if err != nil {
		return videoDetails{}, err
	}

	var resp videoDetails
	err = json.Unmarshal(out, &resp)
	if err != nil {
		logrus.Warnf("Error unmarshalling youtube-dl json output for %s", id)
	}
	return resp, err
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

func cmd(id, formatID string) *exec.Cmd {
	cmd := fmt.Sprintf("youtube-dl -f %s -o - \"%s\" | ffmpeg -i - -f s16le -ar 48000 -ac 2 pipe:1", formatID, id)
	return exec.Command("sh", "-c", cmd)
}
