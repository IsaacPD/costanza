package player

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

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
func getQueue(s *discordgo.Session, key string) (*Queue, bool) {
	_, ok := queueMap[key]
	if !ok && s != nil {
		queueMap[key] = NewQueue(s, key)
	}
	return queueMap[key], ok
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

func Play(s *discordgo.Session, m *discordgo.MessageCreate) {
	queue, _ := getQueue(s, m.GuildID)
	query := strings.SplitN(m.Content, " ", 2)[1]

	if track := autil.GetTrack(query); track != nil {
		queue.InsertTrack(track)
		queue.Play(m.Author.ID)
	} else {
		results := youtube.Search(query)
		lim := util.Min(5, len(results))
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Results:\n%s", youtube.YTTracks(results[:lim])))
	}
}

func PrintQueue(channelID, guildID string) {
	queue, ok := getQueue(nil, guildID)
	if !ok {
		return
	}
	queue.Session.ChannelMessageSend(channelID, queue.String())
}

func QueueTrack(s *discordgo.Session, m *discordgo.MessageCreate) {
	queue, _ := getQueue(s, m.GuildID)
	query := strings.SplitN(m.Content, " ", 2)[1]
	if track := autil.GetTrack(query); track != nil {
		queue.AddTrack(track)
		queue.Play(m.Author.ID)
	} else {
		results := youtube.Search(query)
		lim := util.Min(5, len(results))
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Results:\n%s", youtube.YTTracks(results[:lim])))
	}
}

func Previous(guildID string) {
	queue, ok := getQueue(nil, guildID)
	if !ok {
		return
	}
	queue.Prev()
}

func Skip(s *discordgo.Session, guildID string) {
	queue, ok := getQueue(s, guildID)
	if !ok {
		return
	}
	queue.Skip()
}

func Pause(guildID string) {
	queue, ok := getQueue(nil, guildID)
	if !ok {
		return
	}
	queue.Pause()
}

func UnPause(guildID string) {
	queue, ok := getQueue(nil, guildID)
	if !ok {
		return
	}
	queue.UnPause()
}
