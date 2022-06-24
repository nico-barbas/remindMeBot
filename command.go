package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type (
	command interface {
		getKind() commandKind
		String() string
		execute(u *user) (confirmation *discordgo.MessageEmbed, it *item)
	}

	commandKind int
)

const (
	commandInvalid commandKind = iota
	commandBriefMe
	commandRemindMe
	commandStaffMe
	commandShowTodo
	commandShowReminders
)

var commandKeywords = map[string]commandKind{
	"briefme":    commandBriefMe,
	"remindme":   commandRemindMe,
	"staffme":    commandStaffMe,
	"showtodo":   commandShowTodo,
	"showremind": commandShowReminders,
}

type (
	briefMeCommand struct {
		kind     commandKind
		token    token
		cmdToken token
	}

	remindMeCommand struct {
		kind       commandKind
		token      token
		cmdToken   token
		identifier string
		sepToken   token
		date       date
	}
)

func (b *briefMeCommand) getKind() commandKind { return b.kind }
func (b *briefMeCommand) String() string       { return "Brief me!" }
func (b *briefMeCommand) execute(u *user) (confirmation *discordgo.MessageEmbed, it *item) {
	confirmation = &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: b.String(),
	}
	briefBuilder := strings.Builder{}

	for _, reminder := range u.reminders {
		briefBuilder.WriteString(":small_orange_diamond: **")
		briefBuilder.WriteString(reminder.name)
		briefBuilder.WriteString("**  ||  ")
		briefBuilder.WriteString(reminder.dueTime.Format(timeFormat))
		briefBuilder.WriteString("\n  ")
	}
	confirmation.Fields = append(confirmation.Fields, &discordgo.MessageEmbedField{
		Name:  fmt.Sprintf("%s **Reminders:**", bellEmote),
		Value: strings.Clone(briefBuilder.String()),
	})
	briefBuilder.Reset()

	if len(u.tasks) > 0 {
		for _, task := range u.tasks {
			if task.done {
				briefBuilder.WriteString(todoCheckEmote)
			} else {
				briefBuilder.WriteString(todoUncheckEmote)
			}
			briefBuilder.WriteString(" **")
			briefBuilder.WriteString(task.name)
			briefBuilder.WriteString("**\n  ")
		}
		confirmation.Fields = append(confirmation.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s **Tasks:**", todoEmote),
			Value: strings.Clone(briefBuilder.String()),
		})
	} else {
		confirmation.Fields = append(confirmation.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s **Tasks:**", todoEmote),
			Value: "No active tasks",
		})
	}
	briefBuilder.Reset()
	return
}

func (r *remindMeCommand) getKind() commandKind { return r.kind }
func (r *remindMeCommand) String() string       { return "Remind me!" }
func (r *remindMeCommand) execute(u *user) (confirmation *discordgo.MessageEmbed, it *item) {
	u.reminders = append(u.reminders, item{
		id:   theApp.genTaskID(),
		name: r.identifier,
		kind: itemReminder,
		dueTime: time.Date(
			r.date.year, r.date.month, r.date.day,
			r.date.hour, r.date.min, 0, 0, time.Local,
		),
		needReminder: true,
		done:         false,
	})
	it = &u.reminders[len(u.reminders)-1]

	confirmation = &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       r.String(),
		Description: "Reminder has been added",
	}
	return
}
