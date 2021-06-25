package player

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"layeh.com/gopus"

	"github.com/bwmarrin/discordgo"
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/util"
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
			logrus.Debug("PCM channel closed")
			return
		}
		opus, err := encoder.Encode(receive, util.FRAME_SIZE, util.MAX_BYTES)
		if err != nil {
			logrus.Debug("Encoding error,", err)
			return
		}
		if !voice.Ready || voice.OpusSend == nil {
			logrus.Debugf("Discordgo not ready for opus packets. %+v : %+v", voice.Ready, voice.OpusSend)
			return
		}
		voice.OpusSend <- opus
	}
}

func (connection *Connection) Play(track sound.Track) error {
	defer func() {
		switch {
		default:
			connection.trackEnd <- 1
		}
	}()
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
		connection.isPaused = false
		logrus.Debugf("Finished playing %s", track)
	}()
	connection.voiceConnection.Speaking(true)
	defer connection.voiceConnection.Speaking(false)
	if connection.send == nil {
		connection.send = make(chan []int16, 2)
	}
	defer func() {
		close(connection.send)
		connection.send = nil
	}()
	go connection.sendPCM(connection.voiceConnection, connection.send)
	return connection.read(out, track)
}

func (connection *Connection) read(r io.Reader, track sound.Track) error {
	buffer := bufio.NewReaderSize(r, 16384)
	for {
		if connection.isPaused {
			logrus.Debug("Pausing")
			<-connection.unPause
			logrus.Debug("Unpaused")
		}
		if connection.stopRunning {
			logrus.Debug("Stopping")
			track.Stop()
			break
		}
		audioBuffer := make([]int16, util.FRAME_SIZE*util.CHANNELS)
		err := binary.Read(buffer, binary.LittleEndian, &audioBuffer)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			logrus.Debug("End of track")
			return nil
		}
		if err != nil {
			return err
		}

		connection.send <- audioBuffer
	}
	return nil
}

func (connection *Connection) Stop() {
	connection.stopRunning = true
	logrus.Debug("Waiting for track to stop playing")
}
