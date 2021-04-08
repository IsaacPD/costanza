package player

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	GuildID string
	Session *discordgo.Session

	connection   *Connection
	tracks       []sound.Track
	currentTrack int
	isPaused     bool
}

func NewQueue(s *discordgo.Session, guildID string) *Queue {
	return &Queue{
		GuildID:      guildID,
		Session:      s,
		currentTrack: -1,
	}
}

func (q *Queue) loadNextTrack() {
	if len(q.tracks) == 0 || q.currentTrack+1 >= len(q.tracks) || q.connection.playing {
		return
	}
	q.currentTrack++
	track := q.tracks[q.currentTrack]
	logrus.Debugf("Now playing %s in %s", track, q.connection.voiceConnection.GuildID)
	err := q.connection.Play(track)
	if err != nil {
		logrus.Warnf("Error playing track %s, err: %s", track, err)
	}
	q.loadNextTrack()
}

func (q *Queue) AddTrack(track sound.Track) {
	q.tracks = append(q.tracks, track)
	logrus.Debugf("Added track %v", q.tracks)
}

func (q *Queue) InsertTrack(track sound.Track) {
	if len(q.tracks) < 2 || q.currentTrack+2 >= len(q.tracks) {
		q.AddTrack(track)
		return
	}
	q.tracks = append(q.tracks[:q.currentTrack+2], q.tracks[q.currentTrack+1:]...)
	q.tracks[q.currentTrack+1] = track
	logrus.Debugf("Added track %s", q.tracks)
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
			unPause:         make(chan interface{}),
			trackEnd:        make(chan interface{}),
		}
	}

	if q.connection.playing {
		logrus.Debugf("Song already playing")
		return
	}
	if q.connection.isPaused {
		logrus.Debugf("Continuing paused song")
		q.UnPause()
		return
	}
	q.loadNextTrack()
}

func (q *Queue) Skip() {
	if q.connection == nil || !q.connection.playing {
		return
	}
	q.connection.Stop()
	q.Play("")
}

func (q *Queue) Prev() {
	if q.connection == nil || !q.connection.playing {
		return
	}
	q.currentTrack--
	q.connection.Stop()
	q.Play("")
}

func (q *Queue) Pause() {
	if q.connection == nil || !q.connection.playing {
		return
	}
	q.isPaused = true
	q.connection.isPaused = true
}

func (q *Queue) UnPause() {
	if q.connection == nil || !q.connection.playing || !q.connection.isPaused {
		return
	}
	q.isPaused = false
	q.connection.isPaused = false
	q.connection.unPause <- 1
}

func (q *Queue) String() string {
	return fmt.Sprintf("%v", q.tracks)
}
