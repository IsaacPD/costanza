package player

import (
	"fmt"
	"strings"
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

func (c *Connection) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "{ sendPcm: %v, stopRunning: %v, playing: %v, isPaused: %v }", c.sendpcm, c.stopRunning, c.playing, c.isPaused)
	return sb.String()
}
