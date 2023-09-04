package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const ArchiveDir = "/home/isaacpd/costanza/out/archived"

type (
	Message struct {
		// The ID of the message.
		ID string `json:"id"`

		// The content of the message.
		Content string `json:"content"`

		// The time at which the messsage was sent.
		// CAUTION: this field may be removed in a
		// future API version; it is safer to calculate
		// the creation time via the ID.
		Timestamp time.Time `json:"timestamp"`

		// The time at which the last edit of the message
		// occurred, if it has been edited.
		EditedTimestamp *time.Time `json:"edited_timestamp"`

		// The roles mentioned in the message.
		MentionRoles []string `json:"mention_roles"`

		// Whether the message is text-to-speech.
		TTS bool `json:"tts"`

		// Whether the message mentions everyone.
		MentionEveryone bool `json:"mention_everyone"`

		// The author of the message. This is not guaranteed to be a
		// valid user (webhook-sent messages do not possess a full author).
		Author *discordgo.User `json:"author"`

		// A list of attachments present in the message.
		Attachments []*discordgo.MessageAttachment `json:"attachments"`

		// A list of users mentioned in the message.
		Mentions []*discordgo.User `json:"mentions"`

		// A list of reactions to the message.
		Reactions []*discordgo.MessageReactions `json:"reactions"`

		// Channels specifically mentioned in this message
		// Not all channel mentions in a message will appear in mention_channels.
		// Only textual channels that are visible to everyone in a lurkable guild will ever be included.
		// Only crossposted messages (via Channel Following) currently include mention_channels at all.
		// If no mentions in the message meet these requirements, this field will not be sent.
		MentionChannels []*discordgo.Channel `json:"mention_channels"`

		// MessageReference contains reference data sent with crossposted or reply messages.
		// This does not contain the reference *to* this message; this is for when *this* message references another.
		// To generate a reference to this message, use (*Message).Reference().
		MessageReference *discordgo.MessageReference `json:"message_reference"`

		// The message associated with the message_reference
		// NOTE: This field is only returned for messages with a type of 19 (REPLY) or 21 (THREAD_STARTER_MESSAGE).
		// If the message is a reply but the referenced_message field is not present,
		// the backend did not attempt to fetch the message that was being replied to, so its state is unknown.
		// If the field exists but is null, the referenced message was deleted.
		ReferencedMessage *discordgo.Message `json:"referenced_message"`

		// An array of Sticker objects, if any were sent.
		StickerItems []*discordgo.Sticker `json:"sticker_items"`
	}

	TextChannel struct {
		// The ID of the channel.
		ID string `json:"id"`

		// The ID of the guild to which the channel belongs, if it is in a guild.
		// Else, this ID is empty (e.g. DM channels).
		GuildID string `json:"guild_id"`

		// The name of the channel.
		Name string `json:"name"`

		// The topic of the channel.
		Topic string `json:"topic"`

		// Messages in the thread
		MessageCount int `json:"message_count"`

		// The messages in the channel.
		Messages []*Message `json:"messages"`
	}
)

func ToMessage(m *discordgo.Message) *Message {
	return &Message{
		ID:                m.ID,
		Content:           m.Content,
		Timestamp:         m.Timestamp,
		EditedTimestamp:   m.EditedTimestamp,
		MentionRoles:      m.MentionRoles,
		TTS:               m.TTS,
		MentionEveryone:   m.MentionEveryone,
		Author:            m.Author,
		Attachments:       m.Attachments,
		Mentions:          m.Mentions,
		Reactions:         m.Reactions,
		MentionChannels:   m.MentionChannels,
		MessageReference:  m.MessageReference,
		ReferencedMessage: m.ReferencedMessage,
		StickerItems:      m.StickerItems,
	}
}

func DownloadAttachments(parentDir string, attachments []*discordgo.MessageAttachment) []error {
	var wg sync.WaitGroup
	wg.Add(len(attachments))
	errors := make([]error, len(attachments))
	for i, a := range attachments {
		index := i
		attachment := a
		go func() {
			defer wg.Done()
			req := fasthttp.AcquireRequest()
			res := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseRequest(req)
			defer fasthttp.ReleaseResponse(res)

			req.SetRequestURI(attachment.URL)
			err := DoWithRedirects(req, res)
			if err != nil {
				errors[index] = err
			} else {
				_ = os.Mkdir(fmt.Sprintf("%s/attachments", parentDir), 0700)
				path := fmt.Sprintf("%s/attachments/%s", parentDir, attachment.Filename)
				f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0700)
				body := res.Body()
				_, err := fmt.Fprint(f, string(body))
				if err != nil {
					errors[index] = err
				}
				attachment.Filename = path
			}
		}()
	}
	wg.Wait()
	return errors
}

func Archive(c cmd.Context) (string, error) {
	channelID := c.Args[0]

	c.Defer()
	channel, err := c.Session.Channel(channelID)

	if err != nil {
		c.Followup(fmt.Sprintf("Error reading channel error: %v", err))
		return "", nil
	}

	var allMessages []*discordgo.Message
	for {
		beforeID := ""
		if len(allMessages) > 0 {
			beforeID = allMessages[len(allMessages)-1].ID
		}
		messages, err := c.Session.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			logrus.Warnf("Error reading channel messages channelID: %v, err: %v", channelID, err)
			break
		}
		if len(messages) == 0 {
			break
		}
		allMessages = append(allMessages, messages...)
	}

	logrus.Infof("Read {%d} messages from %s", len(allMessages), channelID)
	logrus.Infof("First Message Time {%v}", allMessages[0].Timestamp)
	logrus.Infof("Last Message Time {%v}", allMessages[len(allMessages)-1].Timestamp)

	archivedChannel := TextChannel{
		ID:           channel.ID,
		GuildID:      channel.GuildID,
		Name:         channel.Name,
		Topic:        channel.Topic,
		MessageCount: len(allMessages),
	}

	// Open directory for archiving
	channelDirStr := fmt.Sprintf("%s/%s", ArchiveDir, strings.ReplaceAll(archivedChannel.Name, " ", "_"))
	err = os.Mkdir(channelDirStr, 0700)
	if err != nil {
		c.Followup("Error creating directory for writing")
		logrus.Warn(err)
		return "", nil
	}

	// Save images and remove unnecessary fields.
	var attachments []*discordgo.MessageAttachment
	for _, message := range allMessages {
		archivedChannel.Messages = append(archivedChannel.Messages, ToMessage(message))
		if len(message.Attachments) > 0 {
			attachments = append(attachments, archivedChannel.Messages[len(archivedChannel.Messages)-1].Attachments...)
		}
	}
	errors := DownloadAttachments(channelDirStr, attachments)
	for _, e := range errors {
		if e != nil {
			logrus.Warn(err)
		}
	}

	// Save messages as json.
	channelJSON, err := json.Marshal(archivedChannel)
	if err != nil {
		c.Followup("Error converting archiving thread to JSON format.")
		logrus.Warn(err)
		return "", nil
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/messages.json", channelDirStr), os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		c.Followup("Error opening file for writing")
		logrus.Warn(err)
		return "", nil
	}
	_, err = fmt.Fprint(file, string(channelJSON))
	logrus.Trace(err)
	c.Followup(fmt.Sprintf("Successfully archived {%d} messages from %s", len(allMessages), channel.Name))
	return "", nil
}
