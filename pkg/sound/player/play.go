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
func getQueue(c cmd.Context) *Queue {
	_, ok := queueMap[c.GuildID]
	if !ok {
		queueMap[c.GuildID] = NewQueue(c)
	}
	return queueMap[c.GuildID]
}

func getChannelWithUser(s *discordgo.Session, userID, guildID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		logrus.Errorf("Error getting channels for guild: %s", err)
		return ""
	}
	var vc string
	for _, state := range guild.VoiceStates {
		if state.UserID == userID {
			vc = state.ChannelID
			break
		}
	}
	return vc
}

func markComplete(c cmd.Context) {
	if c.Message == nil {
		return
	}
	logrus.Tracef("Reacted Error :%s", c.Session.MessageReactionAdd(c.ChannelID, c.Interaction.ID, util.COMPLETE))
}

func Play(c cmd.Context) (string, error) {
	queue := getQueue(c)
	if track := autil.GetTrack(c.Arg); track != nil {
		queue.InsertTrack(track[0])
		queue.Play(c.Author.ID)
		return fmt.Sprintf("Playing next: %s", track[0]), nil
	} else {
		results := youtube.Search(c.Arg)
		lim := util.Min(5, len(results))
		return fmt.Sprintf("Results:\n%s", sound.TrackList(results[:lim])), nil
	}
}

func PrintQueue(c cmd.Context) (string, error) {
	q := getQueue(c)
	if c.Message != nil {
		m, err := c.Session.ChannelMessageSend(c.ChannelID, q.GetPage(0))
		c.Log(m, err)
		_ = c.Session.MessageReactionAdd(c.ChannelID, m.ID, PAGE_PREV)
		_ = c.Session.MessageReactionAdd(c.ChannelID, m.ID, PAGE_NEXT)
		return "", nil
	} else {
		return q.GetPage(0), nil
	}
}

func QueueTrack(c cmd.Context) (string, error) {
	if c.Arg == "" {
		return UnPause(c)
	}
	queue := getQueue(c)
	if tracks := autil.GetTrack(c.Arg); tracks != nil {
		queue.AddTracks(tracks)
		queue.Play(c.Author.ID)
		return fmt.Sprintf("Added %d songs to the queue: %s", len(tracks), tracks), nil
	} else {
		results := youtube.Search(c.Arg)
		lim := util.Min(5, len(results))
		return fmt.Sprintf("Results:\n%s", results[:lim]), nil
	}
}

func Previous(c cmd.Context) (string, error) {
	q := getQueue(c)
	q.Prev()
	markComplete(c)
	if len(q.tracks) == 0 {
		return "No track to go back to.", nil
	}
	return fmt.Sprintf("Playing: %s", q.tracks[q.currentTrack+1]), nil
}

func Skip(c cmd.Context) (string, error) {
	q := getQueue(c)
	q.Skip()
	markComplete(c)
	return "Skipped the song.", nil
}

func Pause(c cmd.Context) (string, error) {
	q := getQueue(c)
	q.Pause()
	markComplete(c)
	if len(q.tracks) == 0 {
		return "No tracks to pause", nil
	}
	return fmt.Sprintf("Pausing: %s", q.tracks[q.currentTrack]), nil
}

func UnPause(c cmd.Context) (string, error) {
	q := getQueue(c)
	q.UnPause()
	markComplete(c)
	if len(q.tracks) == 0 {
		return "No tracks to resume", nil
	}
	return fmt.Sprintf("Resuming: %s", q.tracks[q.currentTrack]), nil
}

func Debug(c cmd.Context) (string, error) {
	q := getQueue(c)
	return fmt.Sprintf("```go\nqueue: {%+v}\nconnection: {%+v}\n```", *q, q.connection), nil
}

func ListDir(c cmd.Context) (string, error) {
	var out []byte
	if c.Arg == "" {
		out, _ = exec.Command("tree", "-d", autil.Music).Output()
	} else {
		path := fmt.Sprintf("%s/%s", autil.Music, c.Arg)
		out, _ = exec.Command("tree", path).Output()
	}
	return string(out), nil
}

func changePage(c cmd.Context, m *discordgo.Message, emoji string) {
	q := getQueue(c)
	pageNum := q.GetPageNum(m.Content)
	pageNum += emojiMap[emoji]
	c.Session.ChannelMessageEdit(m.ChannelID, m.ID, q.GetPage(pageNum))
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
