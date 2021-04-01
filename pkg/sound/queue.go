package sound

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	GuildID string
	Session *discordgo.Session

	voiceConnection *discordgo.VoiceConnection
	tracks          []string
	playing         string
	currentTrack    int
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
	track += ""
}

func (q *Queue) AddTrack(track string) {
	q.tracks = append(q.tracks, track)
}

func (q *Queue) Play() {
	if q.voiceConnection == nil {
		vc, err := connectToFirstVoiceChannel(q.Session, q.GuildID)
		if err != nil {
			logrus.Errorf("Error joining voice channel: %s", err)
		}
		q.voiceConnection = vc
	}

	if q.playing != "" {
		return
	}
	q.loadNextTrack()
}
