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
	queueMap  map[string]*Queue
	PAGE_NEXT = "🔼"
	PAGE_PREV = "🔽"

	emojiMap = map[string]int{
		"🔼": 1,
		"🔽": -1,
		"⬆": 1,
		"⬇": -1,
	}
)

func init() {
	queueMap = make(map[string]*Queue)
}

// getQueue gets the queue for the given key, and creates it if doesn't exist
// returns the queue for the corresponding key and whether it was initially present or not
func getQueue(c cmd.Context, callback func(q *Queue)) {
	_, ok := queueMap[c.GuildID]
	if !ok {
		queueMap[c.GuildID] = NewQueue(c)
	}
	callback(queueMap[c.GuildID])
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
	logrus.Tracef("Reacted Error :%s", c.Session.MessageReactionAdd(c.ChannelID, c.Interaction.ID, util.COMPLETE))
}

func Play(c cmd.Context) {
	getQueue(c, func(queue *Queue) {
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
	getQueue(c, func(q *Queue) {
		m, err := c.Session.ChannelMessageSend(c.ChannelID, q.GetPage(0))
		c.Log(m, err)
		_ = c.Session.MessageReactionAdd(c.ChannelID, m.ID, PAGE_PREV)
		_ = c.Session.MessageReactionAdd(c.ChannelID, m.ID, PAGE_NEXT)
	})
}

func QueueTrack(c cmd.Context) {
	if c.Arg == "" {
		UnPause(c)
		return
	}
	getQueue(c, func(queue *Queue) {
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
	getQueue(c, func(q *Queue) {
		q.Prev()
		markComplete(c)
	})
}

func Skip(c cmd.Context) {
	getQueue(c, func(q *Queue) {
		q.Skip()
		markComplete(c)
	})
}

func Pause(c cmd.Context) {
	getQueue(c, func(q *Queue) {
		q.Pause()
		markComplete(c)
	})
}

func UnPause(c cmd.Context) {
	getQueue(c, func(q *Queue) {
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

func changePage(c cmd.Context, m *discordgo.Message, emoji string) {
	getQueue(c, func(q *Queue) {
		pageNum := q.GetPageNum(m.Content)
		pageNum += emojiMap[emoji]
		c.Session.ChannelMessageEdit(m.ChannelID, m.ID, q.GetPage(pageNum))
	})
}

func reactHandler(s *discordgo.Session, m *discordgo.MessageReaction) {
	if m.UserID == s.State.User.ID {
		return
	}
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		logrus.Warnf("error getting reacted message %s", err)
		return
	}
	if _, ok := emojiMap[m.Emoji.Name]; !ok || message.Author.ID != s.State.User.ID {
		return
	}

	var isPaginated bool
	for _, reaction := range message.Reactions {
		isPaginated = isPaginated || reaction.Me
	}
	if !isPaginated {
		return
	}

	ctx := cmd.Context{Session: s, GuildID: m.GuildID}
	changePage(ctx, message, m.Emoji.Name)
}

func MessageReact(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	reactHandler(s, m.MessageReaction)
}

func MessageRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	reactHandler(s, m.MessageReaction)
}
