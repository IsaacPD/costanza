package sound

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

var (
	queueMap map[string]*Queue
	Music    = "E:\\Desktop\\Stuffs\\Music"
)

func init() {
	queueMap = make(map[string]*Queue)
}

func connectToFirstVoiceChannel(s *discordgo.Session, guildID string) (*discordgo.VoiceConnection, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		logrus.Errorf("Error getting channels for guild: %s", err)
		return nil, err
	}
	var vc string
	for _, c := range channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			vc = c.ID
			break
		}
	}
	return s.ChannelVoiceJoin(guildID, vc, false, false)
}

func Play(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, ok := queueMap[m.GuildID]
	if !ok {
		queueMap[m.GuildID] = NewQueue(s, m.GuildID)
	}
	queue := queueMap[m.GuildID]

	trackName := strings.Split(m.Content, " ")[1]
	if strings.Contains(trackName, "path") {
		trackName = strings.Replace(trackName, "path", Music, 1)
	}
	queue.AddTrack(trackName)
	queue.Play()
}
