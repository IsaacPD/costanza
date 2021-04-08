package util

import (
	"strings"

	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/sound/sources/local"
	"github.com/isaacpd/costanza/pkg/sound/sources/youtube"
)

const Music = "/mnt/e/Desktop/Stuffs/Music"

func getTrackPath(track string) string {
	if strings.Contains(track, "path") {
		return strings.Replace(track, "path", Music, 1)
	}
	return track
}

func GetTrack(track string) sound.Track {
	if path := getTrackPath(track); path != track {
		return local.NewLocalTrack(path)
	}

	return youtube.NewYoutubeTrack(track)
}
