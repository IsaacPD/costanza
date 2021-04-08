package player

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/util"
	"github.com/sirupsen/logrus"
	"layeh.com/gopus"
)

func (connection *Connection) sendPCM(voice *discordgo.VoiceConnection, pcm <-chan []int16) {
	connection.lock.Lock()
	if connection.sendpcm || pcm == nil {
		connection.lock.Unlock()
		return
	}
	connection.sendpcm = true
	connection.lock.Unlock()
	defer func() {
		connection.sendpcm = false
	}()
	encoder, err := gopus.NewEncoder(util.FRAME_RATE, util.CHANNELS, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder error,", err)
		return
	}
	for {
		receive, ok := <-pcm
		if !ok {
			fmt.Println("PCM channel closed")
			return
		}
		opus, err := encoder.Encode(receive, util.FRAME_SIZE, util.MAX_BYTES)
		if err != nil {
			fmt.Println("Encoding error,", err)
			return
		}
		if !voice.Ready || voice.OpusSend == nil {
			fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", voice.Ready, voice.OpusSend)
			return
		}
		voice.OpusSend <- opus
	}
}

func (connection *Connection) Play(track sound.Track) error {
	if connection.playing {
		return errors.New("song already playing")
	}
	connection.stopRunning = false
	out, err := track.GetReader()
	if err != nil {
		return err
	}
	err = track.Start()
	if err != nil {
		return err
	}
	connection.playing = true
	connection.isPaused = false
	defer func() {
		connection.playing = false
		switch {
		default:
			connection.trackEnd <- 1
		}
	}()
	connection.voiceConnection.Speaking(true)
	defer connection.voiceConnection.Speaking(false)
	if connection.send == nil {
		connection.send = make(chan []int16, 2)
	}
	go connection.sendPCM(connection.voiceConnection, connection.send)
	return connection.read(out, track)
}

func (connection *Connection) read(r io.Reader, track sound.Track) error {
	buffer := bufio.NewReaderSize(r, 16384)
	for {
		if connection.isPaused {
			<-connection.unPause
		}
		if connection.stopRunning {
			track.Stop()
			break
		}
		audioBuffer := make([]int16, util.FRAME_SIZE*util.CHANNELS)
		err := binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
		connection.send <- audioBuffer
	}
	return nil
}

func (connection *Connection) Stop() bool {
	connection.stopRunning = true
	connection.playing = false
	connection.isPaused = false
	logrus.Debug("Waiting for track to stop playing")
	<-connection.trackEnd
	return true
}
