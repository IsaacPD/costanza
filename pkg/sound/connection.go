package sound

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Connection struct {
	voiceConnection *discordgo.VoiceConnection
	send            chan []int16
	lock            sync.Mutex
	sendpcm         bool
	stopRunning     bool
	playing         bool
}
