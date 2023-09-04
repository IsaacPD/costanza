package player

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/util"
)

const PAGE_SIZE = 5

type Queue struct {
	*cmd.Context

	connection   *Connection
	tracks       sound.TrackList
	currentTrack int
	isPaused     bool
}

func NewQueue(c cmd.Context) *Queue {
	return &Queue{
		Context:      &c,
		currentTrack: -1,
	}
}

func (q *Queue) loadNextTrack() {
	if len(q.tracks) == 0 || q.currentTrack+1 >= len(q.tracks) || q.connection.playing {
		return
	}
	q.currentTrack++
	track := q.tracks[q.currentTrack]
	logrus.Debugf("Now playing: %s in %s", track, q.connection.voiceConnection.GuildID)
	q.Send(fmt.Sprintf("Now playing %s", track))
	err := q.connection.Play(track)
	if err != nil {
		logrus.Warnf("Error playing track %s, err: %s", track, err)
	}
}

func (q *Queue) rotateQueue() {
	for {
		select {
		case <-q.connection.trackEnd:
			go q.loadNextTrack()
		case <-q.connection.endConnection:
			logrus.Debug("Connection ended")
			return
		}
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
	vcID := getChannelWithUser(q.Session, userID, q.GuildID)
	if vcID == "" {
		q.Send("Please join a voice channel to make a request")
		return
	}

	if q.connection == nil || vcID != q.connection.voiceConnection.ChannelID {
		vc, err := q.Session.ChannelVoiceJoin(q.GuildID, vcID, false, false)
		if err != nil {
			logrus.Errorf("Error joining voice channel: %s", err)
			return
		}
		if q.connection != nil {
			logrus.Debug("Sending request to end connection")
			q.connection.endConnection <- 1
			q.connection.Stop()
			// <-q.connection.pcmClosed
			q.connection.voiceConnection = vc
			time.Sleep(1 * time.Second)
		} else {
			q.connection = &Connection{
				voiceConnection: vc,
				unPause:         make(chan interface{}),
				trackEnd:        make(chan interface{}),
				endConnection:   make(chan interface{}),
				pcmClosed:       make(chan interface{}),
			}
		}
		logrus.Tracef("Starting queue rotation")
		go q.rotateQueue()
	}
	_, err := q.Session.State.VoiceState(q.GuildID, q.Session.State.User.ID)
	if err != nil {
		logrus.Warnf("Currently not in a voice connection: %v. Attempting to rejoin", err)
		_, err = q.Session.ChannelVoiceJoin(q.GuildID, vcID, false, false)
		if err != nil {
			logrus.Errorf("Could not rejoin voice channel %s", err)
			return
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
	logrus.Tracef("Triggering start of next track")
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

func (*Queue) GetPageNum(text string) int {
	lines := strings.Split(text, "\n")
	last := lines[len(lines)-2]

	var i int
	_, err := fmt.Sscanf(last, "%d)", &i)
	if err != nil {
		return -1
	}
	i--

	return i / PAGE_SIZE
}

func (q *Queue) GetPage(num int) string {
	start := PAGE_SIZE * num
	if start >= len(q.tracks) {
		start = 0
	}
	if start < 0 {
		start = ((len(q.tracks) - 1) / PAGE_SIZE) * PAGE_SIZE
	}
	end := start + PAGE_SIZE

	return q.tracks[start:util.Min(end, len(q.tracks))].StringWithOffset(start)
}
