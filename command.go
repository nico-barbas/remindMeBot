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
	commandRemoveMe
	commandShowTodo
	commandShowReminders
)

var commandKeywords = map[string]commandKind{
	"briefme":    commandBriefMe,
	"remindme":   commandRemindMe,
	"staffme":    commandStaffMe,
	"removeme":   commandRemoveMe,
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

	staffMeCommand struct {
		kind       commandKind
		token      token
		cmdToken   token
		identifier string
		sepToken   token
		hasDueDate bool
		date       date
	}

	removeMeCommand struct {
		kind       commandKind
		token      token
		cmdToken   token
		list       token
		sepToken   token
		identifier string
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

	if len(u.reminders) > 0 {
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
	} else {
		confirmation.Fields = append(confirmation.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s **Reminders:**", bellEmote),
			Value: "No active reminders",
		})
	}
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
		id:         theApp.genItemID(),
		name:       r.identifier,
		kind:       itemReminder,
		hasDueDate: true,
		dueTime: time.Date(
			r.date.year, r.date.month, r.date.day,
			r.date.hour, r.date.min, 0, 0, time.Local,
		),
		alarmCount: 0,
		done:       false,
	})
	it = &u.reminders[len(u.reminders)-1]

	confirmation = &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       r.String(),
		Description: "Reminder has been added",
	}
	return
}

func (s *staffMeCommand) getKind() commandKind { return s.kind }
func (s *staffMeCommand) String() string       { return "Staff me!" }
func (s *staffMeCommand) execute(u *user) (confirmation *discordgo.MessageEmbed, it *item) {
	u.tasks = append(u.tasks, item{
		id:         theApp.genItemID(),
		name:       s.identifier,
		kind:       itemTask,
		hasDueDate: s.hasDueDate,
		done:       false,
	})
	it = &u.tasks[len(u.tasks)-1]
	if s.hasDueDate {
		it.dueTime = time.Date(
			s.date.year, s.date.month, s.date.day,
			s.date.hour, s.date.min, 0, 0, time.Local,
		)
	}

	confirmation = &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       s.String(),
		Description: "Task has been added",
	}
	return
}

func (r *removeMeCommand) getKind() commandKind { return r.kind }
func (r *removeMeCommand) String() string       { return "Staff me!" }
func (r *removeMeCommand) execute(u *user) (confirmation *discordgo.MessageEmbed, it *item) {
	found := false
	removed := &item{}
	switch r.list.kind {
	case tokenReminder:
		for i, item := range u.reminders {
			if item.name == r.identifier {
				found = true
				*removed = item
				if len(u.reminders) > 1 {
					copy(u.reminders[i-1:], u.reminders[i:])
					u.reminders = u.reminders[:len(u.reminders)-1]
				} else {
					u.reminders = u.reminders[:0]
				}
				break
			}
		}
	case tokenTask:
		for i, item := range u.tasks {
			if item.name == r.identifier {
				found = true
				*removed = item
				if len(u.tasks) > 1 {
					copy(u.tasks[i-1:], u.tasks[i:])
					u.tasks = u.tasks[:len(u.tasks)-1]
				} else {
					u.tasks = u.tasks[:0]
				}
				break
			}
		}
	}

	listName := "reminder"
	if r.list.kind == tokenTask {
		listName = "task"
	}

	confirmation = &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: r.String(),
	}
	if found {
		confirmation.Description = fmt.Sprintf("%s has been removed", listName)
		it = removed
	} else {
		confirmation.Description = fmt.Sprintf("%s %s does not exist", listName, r.identifier)
	}
	return
}
