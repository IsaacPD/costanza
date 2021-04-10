package player

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/isaacpd/costanza/pkg/sound"
)

type Queue struct {
	cmd.Context

	connection   *Connection
	tracks       sound.TrackList
	currentTrack int
	isPaused     bool
}

func NewQueue(c cmd.Context, guildID string) *Queue {
	return &Queue{
		Context:      c,
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
	q.Send(fmt.Sprintf("Now playing %s", track))
	err := q.connection.Play(track)
	if err != nil {
		logrus.Warnf("Error playing track %s, err: %s", track, err)
	}
}

func (q *Queue) rotateQueue() {
	for {
		<-q.connection.trackEnd
		go q.loadNextTrack()
	}
}

func (q *Queue) AddTracks(track sound.TrackList) {
	q.tracks = append(q.tracks, track...)
	logrus.Tracef("Added tracks %v", track)
	if len(track) > 1 {
		q.Send(fmt.Sprintf("Queued %d tracks", len(track)))
	}
}

func (q *Queue) InsertTrack(track sound.Track) {
	if len(q.tracks) < 2 || q.currentTrack+2 >= len(q.tracks) {
		q.AddTracks([]sound.Track{track})
		return
	}
	q.tracks = append(q.tracks[:q.currentTrack+2], q.tracks[q.currentTrack+1:]...)
	q.tracks[q.currentTrack+1] = track
	logrus.Tracef("Added track %s", track)
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
		go q.rotateQueue()
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
	q.connection.trackEnd <- 1
}

func (q *Queue) Skip() {
	if q.connection == nil || !q.connection.playing {
		return
	}
	q.connection.Stop()
}

func (q *Queue) Prev() {
	if q.connection == nil || q.currentTrack-2 < -1 {
		return
	}
	q.currentTrack -= 2
	q.connection.Stop()
	if !q.connection.playing {
		q.connection.trackEnd <- 1
	}
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
	return q.tracks.String()
}
