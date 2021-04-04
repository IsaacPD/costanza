package sound

import (
	"os/exec"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	GuildID string
	Session *discordgo.Session

	connection   *Connection
	tracks       []string
	playing      string
	currentTrack int
}

func Ffmpeg(song string) *exec.Cmd {
	return exec.Command("ffmpeg", "-i", song, "-f", "s16le", "-ar", strconv.Itoa(FRAME_RATE), "-ac",
		strconv.Itoa(CHANNELS), "pipe:1")
}

func NewQueue(s *discordgo.Session, guildID string) *Queue {
	return &Queue{
		GuildID:      guildID,
		Session:      s,
		currentTrack: -1,
	}
}

func (q *Queue) loadNextTrack() {
	if len(q.tracks) == 0 || q.currentTrack+1 >= len(q.tracks) {
		return
	}
	q.currentTrack++
	track := q.tracks[q.currentTrack]
	err := q.connection.Play(Ffmpeg(track))
	if err != nil {
		logrus.Warnf("Error playing track %s, err: %s", track, err)
	}
}

func (q *Queue) AddTrack(track string) {
	q.tracks = append(q.tracks, track)
}

// Play establishes a connection in the channel where userID if it does
// not exist and loads the next track to be played.
func (q *Queue) Play(userID string) {
	if q.connection == nil {
		vc, err := connectToFirstVoiceChannel(q.Session, userID, q.GuildID)
		if err != nil {
			logrus.Errorf("Error joining voice channel: %s", err)
			return
		}
		q.connection = &Connection{
			voiceConnection: vc,
		}
	}

	if q.playing != "" {
		return
	}
	q.loadNextTrack()
}
