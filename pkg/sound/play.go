package sound

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var (
	queueMap map[string]*Queue
	Music    = "/mnt/e/Desktop/Stuffs/Music"
)

func init() {
	queueMap = make(map[string]*Queue)
}

// getQueue gets the queue for the given key, and creats it if doesn't exist
// returns the queue for the corresponding key and whether it was initially present or not
func getQueue(s *discordgo.Session, key string) (*Queue, bool) {
	_, ok := queueMap[key]
	if !ok && s != nil {
		queueMap[key] = NewQueue(s, key)
	}
	return queueMap[key], ok
}

func getTrackPath(track string) string {
	if strings.Contains(track, "path") {
		return strings.Replace(track, "path", Music, 1)
	}
	return track
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
	queue.InsertTrack(getTrackPath(strings.Split(m.Content, " ")[1]))
	queue.Play(m.Author.ID)
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
	queue.AddTrack(getTrackPath(strings.Split(m.Content, " ")[1]))
	queue.Play(m.Author.ID)
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
