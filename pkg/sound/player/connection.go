package player

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
	isPaused        bool
	pcmClosed       chan interface{}
	endConnection   chan interface{}
	unPause         chan interface{}
	trackEnd        chan interface{}
}
