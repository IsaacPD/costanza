package util

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/sirupsen/logrus"
)

type (
	Vote struct {
		count uint
		user  *discordgo.User
	}
	Votes struct {
		total uint
		votes map[string]*Vote
	}
	Poll struct {
		*cmd.Context

		id string

		prompt  string
		order   []string
		choices map[string]*Votes
		voters  map[string]map[string]uint

		description     string
		allow_change    bool
		bounded         bool
		anon            bool
		live            bool
		select_multiple bool
		end_time        time.Time
	}
)

var (
	pollMap     = make(map[string]*Poll)
	OptionEmoji []rune
	CloseButton = regexp.MustCompile(".*;close")
)

func init() {
	start := '🇦'
	end := '🇿'
	for i := start; i <= end; i++ {
		OptionEmoji = append(OptionEmoji, i)
	}
}

func hiddenFollowup(c cmd.Context, message string) {
	err := c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logrus.Warnf("error sending followup")
	}
}

func (p *Poll) FinalizeDiscordPoll() []*discordgo.MessageEmbed {
	var choices []*discordgo.MessageEmbedField
	index := 0
	for _, c := range p.order {
		count := p.choices[c]
		val := fmt.Sprintf("Total: %d", count.total)
		if !p.anon {
			for _, vote := range count.votes {
				if vote.count == 0 {
					continue
				}
				val += fmt.Sprintf(", %s: %d", vote.user.Username, vote.count)
			}
		}
		choices = append(choices, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%c %s", OptionEmoji[index], c),
			Value: val,
		})
		index++
	}
	return []*discordgo.MessageEmbed{{
		Type:        discordgo.EmbedTypeRich,
		Title:       p.prompt,
		Description: p.description,
		Color:       0xff6a6a,
		Fields:      choices,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Poll by %s\nClosing in %s", p.Author.Username, time.Until(p.end_time).Truncate(time.Minute)),
		},
	}}
}

func (p *Poll) CreateDiscordPoll() ([]*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	var choices []*discordgo.MessageEmbedField
	var buttons []discordgo.MessageComponent
	index := 0
	for _, c := range p.order {
		count := p.choices[c]
		var val string
		if p.live {
			val = fmt.Sprintf("Total: %d", count.total)
		}
		choices = append(choices, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%c %s", OptionEmoji[index], c),
			Value: val,
		})
		buttons = append(buttons, discordgo.Button{
			CustomID: fmt.Sprintf("%s,%s", p.id, c),
			Style:    discordgo.PrimaryButton,
			Label:    c,
			Emoji: discordgo.ComponentEmoji{
				Name: string(OptionEmoji[index]),
			},
		})
		index++
	}
	buttons = append(buttons, discordgo.Button{
		CustomID: fmt.Sprintf("%s;close", p.id),
		Style:    discordgo.DangerButton,
		Label:    "Close Poll",
	})

	return []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       p.prompt,
			Description: p.description,
			Color:       0xff6a6a,
			Fields:      choices,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Poll by %s\nClosing in %s", p.Author.Username, time.Until(p.end_time).Truncate(time.Minute)),
			},
		}}, []discordgo.MessageComponent{&discordgo.ActionsRow{
			Components: buttons,
		}}
}

func ClosePoll(c cmd.Context, poll *Poll) {
	embeds := poll.FinalizeDiscordPoll()
	err := c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
		},
	})
	if err != nil {
		logrus.Warnf("Error updating poll: %v", err)
	}
	delete(pollMap, poll.id)
}

func AltHandlePoll(c cmd.Context) {
	data := c.Interaction.MessageComponentData()
	logrus.Debugf("Attempting to handle poll: %+v", data)

	var id, option string
	var close bool
	if close = CloseButton.MatchString(data.CustomID); close {
		id = data.CustomID[0:strings.Index(data.CustomID, ";")]
	} else {
		out := strings.Split(data.CustomID, ",")
		id = out[0]
		option = out[1]
	}

	poll, ok := pollMap[id]
	if !ok {
		hiddenFollowup(c, "error: this poll is no longer valid")
		return
	}
	if close {
		if c.Author.ID == poll.Author.ID {
			ClosePoll(c, poll)
		} else {
			hiddenFollowup(c, "error: you cannot close the pool")
		}
		return
	}

	userVotes, ok := poll.voters[c.Author.ID]
	if ok && !poll.select_multiple {
		hiddenFollowup(c, "error: this poll does not allow voting again")
		return
	} else {
		poll.voters[c.Author.ID] = make(map[string]uint)
		userVotes = poll.voters[c.Author.ID]
	}

	choice := poll.choices[option]
	vote, ok := choice.votes[c.Author.ID]

	if !ok {
		// New vote so we increment
		vote = &Vote{}
		vote.user = c.Author
		choice.votes[c.Author.ID] = vote
		poll.voters[c.Author.ID][option] = 1
	} else if poll.bounded {
		hiddenFollowup(c, "error: this poll does not allow multiple votes on the same item")
		return
	}
	choice.total++
	vote.count++
	userVotes[option]++

	if !poll.live {
		var sb strings.Builder
		sb.WriteString("{")
		for option, votes := range poll.choices {
			for user, vote := range votes.votes {
				if vote.count == 0 || user != c.Author.ID {
					continue
				}
				fmt.Fprintf(&sb, "{%s: %d}", option, vote.count)
			}
		}
		sb.WriteString("}")
		hiddenFollowup(c, fmt.Sprintf("Your current votes are %s", sb.String()))
	} else {
		embeds, components := poll.CreateDiscordPoll()
		err := c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
		if err != nil {
			logrus.Warnf("Error updating poll: %v", err)
		}
	}
}

func HandlePoll(c cmd.Context) {
	data := c.Interaction.MessageComponentData()
	logrus.Debugf("Attempting to handle poll: %+v", data)

	var id, option string
	var close bool
	if close = CloseButton.MatchString(data.CustomID); close {
		id = data.CustomID[0:strings.Index(data.CustomID, ";")]
	} else {
		out := strings.Split(data.CustomID, ",")
		id = out[0]
		option = out[1]
	}

	poll, ok := pollMap[id]
	if !ok {
		hiddenFollowup(c, "error: this poll is no longer valid")
		return
	}
	if close {
		if c.Author.ID == poll.Author.ID {
			ClosePoll(c, poll)
		} else {
			hiddenFollowup(c, "error: you cannot close the pool")
		}
		return
	}

	userVotes, ok := poll.voters[c.Author.ID]
	if ok && !poll.allow_change {
		hiddenFollowup(c, "error: this poll does not allow changing your vote")
		return
	}
	if ok && !poll.select_multiple {
		// check for previous vote
		for choice := range userVotes {
			if choice == option {
				continue
			}
			opt, ok := poll.choices[choice]
			if ok {
				user, ok := opt.votes[c.Author.ID]
				if !ok || user.count == 0 {
					continue
				}
			}
			opt.total--
			delete(opt.votes, c.Author.ID)
		}
	} else {
		poll.voters[c.Author.ID] = make(map[string]uint)
	}

	choice := poll.choices[option]
	vote, ok := choice.votes[c.Author.ID]

	if !ok || vote.count == 0 {
		// New vote so we increment
		vote = &Vote{}
		vote.user = c.Author
		choice.votes[c.Author.ID] = vote
		choice.total++
		vote.count++
		poll.voters[c.Author.ID][option] = 1
	} else {
		// Existing vote so decrement if we selected it again
		// if poll.bounded {
		// 	hiddenFollowup(c, "error: this poll does not allow multiple votes on the same item")
		// 	return
		// }
		choice.total--
		vote.count--
	}

	if !poll.live {
		var sb strings.Builder
		sb.WriteString("{")
		for option, votes := range poll.choices {
			for user, vote := range votes.votes {
				if vote.count == 0 || user != c.Author.ID {
					continue
				}
				fmt.Fprintf(&sb, "{%s: %d}", option, vote.count)
			}
		}
		sb.WriteString("}")
		hiddenFollowup(c, fmt.Sprintf("Your current votes are %s", sb.String()))
	} else {
		embeds, components := poll.CreateDiscordPoll()
		err := c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     embeds,
				Components: components,
			},
		})
		if err != nil {
			logrus.Warnf("Error updating poll: %v", err)
		}
	}
}

func PollCommand() cmd.Command {
	return cmd.NewCmdWithOptions(cmd.Names{"poll"}, poll, "Creates a poll with the specified options",
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "prompt",
			Description: "the prompt you are polling for",
			Required:    true,
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "choices",
			Description: "Add choices with commas. e.g. 'One, Two, 🥖 Three.'. Put commas within a choice with '\\'",
			Required:    true,
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "description",
			Description: "extra text that will go underneath the prompt",
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "anon",
			Description: "defaults to false. If true then nobody's vote is revealed",
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "select_multiple",
			Description: "defaults to false. Allows selecting more than just one choice",
		},
		&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "live",
			Description: "defaults to true. If true results are updated live otherwise when the poll is closed",
		})
}

func poll(c cmd.Context) (string, error) {
	if c.Interaction == nil {
		return "", errors.New("this command is only supported via slash commands")
	}

	prompt := cmd.GetOption[string](&c, "prompt")
	choices := strings.Split(cmd.GetOption[string](&c, "choices"), ",")

	if len(choices) >= 25 {
		return "", errors.New("too many choices provided, the max is 25")
	}

	time_str := cmd.GetOptionWithDefault[string](&c, "time", "24h")
	duration, err := time.ParseDuration(time_str)
	if err != nil {
		return "", fmt.Errorf("error parsing duration of poll, %v", err)
	}
	end_time := time.Now().Add(duration)

	counts := make(map[string]*Votes)
	var order []string
	for _, i := range choices {
		_, ok := counts[strings.TrimSpace(i)]
		if ok {
			return "", fmt.Errorf("you cannot have multiple of the same choice")
		}
		counts[strings.TrimSpace(i)] = &Votes{votes: make(map[string]*Vote)}
		order = append(order, strings.TrimSpace(i))
	}

	id := uuid.NewString()
	pollMap[id] = &Poll{
		Context:         &c,
		id:              id,
		prompt:          prompt,
		order:           order,
		choices:         counts,
		voters:          make(map[string]map[string]uint),
		description:     cmd.GetOptionWithDefault[string](&c, "description", ""),
		allow_change:    cmd.GetOptionWithDefault[bool](&c, "allow_change", true),
		bounded:         cmd.GetOptionWithDefault[bool](&c, "bounded", true),
		anon:            cmd.GetOptionWithDefault[bool](&c, "anon", false),
		live:            cmd.GetOptionWithDefault[bool](&c, "live", true),
		select_multiple: cmd.GetOptionWithDefault[bool](&c, "select_multiple", false),
		end_time:        end_time,
	}

	embeds, components := pollMap[id].CreateDiscordPoll()

	err = c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Components: components,
		},
	})

	return "", err
}
