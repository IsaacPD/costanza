package player

import (
	"fmt"
	"os/exec"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/sound/sources/youtube"
	autil "github.com/isaacpd/costanza/pkg/sound/util"
	"github.com/isaacpd/costanza/pkg/util"
)

var (
	queueMap map[string]*Queue
)

func init() {
	queueMap = make(map[string]*Queue)
}

// getQueue gets the queue for the given key, and creates it if doesn't exist
// returns the queue for the corresponding key and whether it was initially present or not
func getQueue(c cmd.Context, key string, callback func(q *Queue)) {
	_, ok := queueMap[key]
	if !ok {
		queueMap[key] = NewQueue(c, key)
	}
	callback(queueMap[key])
}

func connectToFirstVoiceChannel(s *discordgo.Session, userID, guildID string) (*discordgo.VoiceConnection, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		logrus.Errorf("Error getting channels for guild: %s", err)
		return nil, err
	}
	var vc string
	for _, state := range guild.VoiceStates {
		if state.UserID == userID {
			vc = state.ChannelID
			break
		}
	}
	logrus.Tracef("Voice Channel to connect to: %s", vc)
	return s.ChannelVoiceJoin(guildID, vc, false, false)
}

func markComplete(c cmd.Context) {
	logrus.Tracef("Reacted Error :%s", c.Session.MessageReactionAdd(c.ChannelID, c.Message.ID, util.COMPLETE))
}

func Play(c cmd.Context) {
	getQueue(c, c.GuildID, func(queue *Queue) {
		if track := autil.GetTrack(c.Arg); track != nil {
			queue.InsertTrack(track[0])
			queue.Play(c.Author.ID)
		} else {
			results := youtube.Search(c.Arg)
			lim := util.Min(5, len(results))
			c.Send(fmt.Sprintf("Results:\n%s", sound.TrackList(results[:lim])))
		}
		markComplete(c)
	})
}

func PrintQueue(c cmd.Context) {
	getQueue(c, c.GuildID, func(q *Queue) {
		c.Send(q.String())
	})
}

func QueueTrack(c cmd.Context) {
	if c.Arg == "" {
		UnPause(c)
		return
	}
	getQueue(c, c.GuildID, func(queue *Queue) {
		if tracks := autil.GetTrack(c.Arg); tracks != nil {
			queue.AddTracks(tracks)
			queue.Play(c.Author.ID)
		} else {
			results := youtube.Search(c.Arg)
			lim := util.Min(5, len(results))
			c.Send(fmt.Sprintf("Results:\n%s", results[:lim]))
		}
		markComplete(c)
	})
}

func Previous(c cmd.Context) {
	getQueue(c, c.GuildID, func(q *Queue) {
		q.Prev()
		markComplete(c)
	})
}

func Skip(c cmd.Context) {
	getQueue(c, c.GuildID, func(q *Queue) {
		q.Skip()
		markComplete(c)
	})
}

func Pause(c cmd.Context) {
	getQueue(c, c.GuildID, func(q *Queue) {
		q.Pause()
		markComplete(c)
	})
}

func UnPause(c cmd.Context) {
	getQueue(c, c.GuildID, func(q *Queue) {
		q.UnPause()
		markComplete(c)
	})
}

func ListDir(c cmd.Context) {
	var out []byte
	if c.Arg == "" {
		out, _ = exec.Command("tree", "-d", autil.Music).Output()
	} else {
		path := fmt.Sprintf("%s/%s", autil.Music, c.Arg)
		out, _ = exec.Command("tree", path).Output()
	}
	c.Send(string(out))
}
