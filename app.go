package main

import (
	"fmt"
	"log"
	"os"
	"remindMeBot/toml"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
)

const (
	sleepTime         = 1 * time.Second
	minBeforeRemind   = 30
	timeFormat        = time.RFC822Z
	initItemBufferCap = 20

	reminderAlarm = "Reminder Notification"
	taskAlarm     = "Task Notification"
)

var theApp *app

type (
	app struct {
		s               *discordgo.Session
		remindChannelID string

		db *genji.DB

		users       map[string]*user
		shouldClose chan bool
		mut         sync.Mutex
		lastTime    time.Time
		config      appConfig
	}

	appConfig struct {
		ItemCounter       int32
		ReminderFrequency int

		AlarmTime struct {
			First  int
			Second int
		}
	}

	user struct {
		uniqueID  int
		id        string
		name      string
		reminders []item
		tasks     []item
	}

	item struct {
		id             int
		name           string
		kind           itemKind
		hasDueDate     bool
		dueTime        time.Time
		alarmCount     int
		lastRemindTime time.Time
		done           bool
	}

	itemKind int
)

const (
	itemInvalid itemKind = iota
	itemReminder
	itemTask
)

func (a *app) init() {
	configFile, _ := os.ReadFile("./data/config.toml")
	toml.Deserialize(string(configFile), &a.config)

	userResults, err := a.db.Query("SELECT id, discord_id, name FROM users;")
	defer userResults.Close()
	if err != nil {
		log.Panicln(err)
	}
	err = userResults.Iterate(func(d types.Document) error {
		var id int
		var discordID string
		var name string

		err = document.Scan(d, &id, &discordID, &name)
		a.users[discordID] = &user{
			uniqueID:  id,
			id:        discordID,
			name:      name,
			reminders: make([]item, 0, initItemBufferCap),
			tasks:     make([]item, 0, initItemBufferCap),
		}
		return err
	})
	if err != nil {
		log.Panicln(err)
	}

	for _, u := range a.users {
		itemResults, err := a.db.Query("SELECT id, name, kind, due_time, done FROM items WHERE user_id = ?;", u.uniqueID)
		defer itemResults.Close()
		if err != nil {
			log.Panicln(err)
		}

		err = itemResults.Iterate(func(d types.Document) error {
			var id int
			var name string
			var kind int
			var dueTimeStr string
			var done int

			err = document.Scan(d, &id, &name, &kind, &dueTimeStr, &done)
			if err != nil {
				return err
			}

			var dueTime time.Time
			hasDueTime := false
			if dueTimeStr != "" {
				dueTime, err = time.Parse(timeFormat, dueTimeStr)
				if err != nil {
					return err
				}
				hasDueTime = true
			}
			newItem := item{
				id:         id,
				name:       name,
				kind:       itemKind(kind),
				hasDueDate: hasDueTime,
				done:       done == 1,
			}
			if hasDueTime {
				newItem.dueTime = dueTime
				remainingTime := int(dueTime.Sub(time.Now()).Minutes())
				if remainingTime <= a.config.AlarmTime.First && remainingTime > a.config.AlarmTime.Second {
					newItem.alarmCount = 1
				} else if remainingTime <= a.config.AlarmTime.Second {
					newItem.alarmCount = 2
				}
			}
			switch itemKind(kind) {
			case itemReminder:
				u.reminders = append(u.reminders, newItem)
			case itemTask:
				u.tasks = append(u.tasks, newItem)
			}
			return err
		})
		if err != nil {
			log.Panicln(err)
		}
	}
}

func (a *app) shutdown() {
	configStr, err := toml.Serialize(&a.config)
	fmt.Println(a.config)
	fmt.Println(configStr)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile("./data/config.toml", []byte(configStr), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func (a *app) run() {
	a.lastTime = time.Now()

runLoop:
	for {
		select {
		case close := <-a.shouldClose:
			if close {
				break runLoop
			}
		default:
			a.mut.Lock()

			for _, user := range a.users {
				a.updateUser(user)
			}

			a.lastTime = time.Now()
			a.mut.Unlock()
			time.Sleep(sleepTime)
		}
	}
}

func (a *app) updateUser(u *user) {
	now := time.Now()
	for i := range u.reminders {
		reminder := &u.reminders[i]
		timeRem := int(reminder.dueTime.Sub(now).Minutes())
		if timeRem < 0 {
			if !reminder.done {
				reminder.lastRemindTime = now
				reminder.done = true
			} else {
				remindRemaining := time.Since(reminder.lastRemindTime).Minutes()
				if remindRemaining >= float64(a.config.ReminderFrequency) {
					reminder.lastRemindTime = now
					a.s.ChannelMessageSend(
						a.remindChannelID,
						fmt.Sprintf("<@%s>", u.id),
					)
					remindMsg, _ := a.s.ChannelMessageSendEmbed(
						a.remindChannelID,
						&discordgo.MessageEmbed{
							Type:        discordgo.EmbedTypeRich,
							Title:       reminderAlarm,
							Description: fmt.Sprintf("Have you done **%s**?", reminder.name),
						},
					)
					a.s.MessageReactionAdd(
						a.remindChannelID,
						remindMsg.ID,
						"☑",
					)
				}
			}
		} else {
			if timeRem <= a.config.AlarmTime.First && timeRem > a.config.AlarmTime.Second {
				if reminder.alarmCount == 0 {
					reminder.alarmCount = 1

					a.s.ChannelMessageSend(
						a.remindChannelID,
						fmt.Sprintf("<@%s>", u.id),
					)
					a.s.ChannelMessageSendEmbed(
						a.remindChannelID,
						&discordgo.MessageEmbed{
							Type:        discordgo.EmbedTypeRich,
							Title:       reminderAlarm,
							Description: fmt.Sprintf("**%s** is in less than 120 minutes (~%d)", reminder.name, timeRem),
						},
					)
				}
			} else if timeRem <= a.config.AlarmTime.Second {
				if reminder.alarmCount < 2 {
					reminder.alarmCount = 2

					a.s.ChannelMessageSend(
						a.remindChannelID,
						fmt.Sprintf("<@%s>", u.id),
					)
					a.s.ChannelMessageSendEmbed(
						a.remindChannelID,
						&discordgo.MessageEmbed{
							Type:        discordgo.EmbedTypeRich,
							Title:       reminderAlarm,
							Description: fmt.Sprintf("**%s** is in less than 30 minutes (~%d)", reminder.name, timeRem),
						},
					)
				}
			}
		}
	}
}

func (a *app) handleError(channelID string, err parserError) {
	a.mut.Lock()
	defer a.mut.Unlock()

	var errString string

	switch err.kind {
	case errorInvalidToken:
		errString = fmt.Sprintf("Invalid token: %s", err.details)

	case errorInvalidSyntax:
		errString = fmt.Sprintf("Invalid syntax: %s", err.details)

	case errorInvalidDate:
		errString = fmt.Sprintf("Invalid date: %s", err.details)

	case errorUnknownCommand:
		errString = fmt.Sprintf("Unknown command: %s", err.details)

	}

	_, msgerr := a.s.ChannelMessageSend(channelID, errString)
	if msgerr != nil {
		log.Println(msgerr)
	}
}

func (a *app) handleCommand(m *discordgo.Message, cmd command) {
	a.mut.Lock()
	defer a.mut.Unlock()

	if _, exist := a.users[m.Author.ID]; !exist {
		// Not in memory, checking DB
		result, err := a.db.Query("SELECT id, discord_id, name FROM users WHERE discord_id = ?;", m.Author.ID)
		if err != nil {
			log.Println(err)
			return
		}

		var count int
		var u *user
		err = result.Iterate(func(d types.Document) error {
			u = &user{}

			err = document.Scan(d, &u.uniqueID, &u.id, &u.name)
			count += 1
			return err
		})
		result.Close()
		if count > 1 {
			log.Printf(
				"Duplicate User in database: id: %d, discordID: %s, name: %s",
				u.uniqueID,
				u.id,
				u.name,
			)
			return
		} else if count == 0 {
			err = a.registerUser(m.Author)
			if err != nil {
				log.Println("DB access failure: ", err)
				return
			}
		}
	}
	user := a.users[m.Author.ID]
	confirmationMsg, it := cmd.execute(user)
	if it != nil {
		switch cmd.(type) {
		case *remindMeCommand, *staffMeCommand:
			dueTimeStr := ""
			if it.hasDueDate {
				dueTimeStr = it.dueTime.Format(timeFormat)
			}
			err := a.db.Exec(
				"INSERT INTO items (id, name, user_id, kind, due_time, done) VALUES (?, ?, ?, ?, ?, ?);",
				it.id,
				it.name,
				user.uniqueID,
				it.kind,
				dueTimeStr,
				it.done,
			)
			if err != nil {
				log.Println("DB access failure: ", err)
			}

		case *removeMeCommand:
			err := a.db.Exec("DELETE FROM items WHERE id = ?;", it.id)
			if err != nil {
				log.Println("DB access failure: ", err)
			}
		}

	}

	_, err := a.s.ChannelMessageSendEmbed(m.ChannelID, confirmationMsg)
	if err != nil {
		log.Println(err)
	}
}

func (a *app) registerUser(u *discordgo.User) error {
	newUser := &user{
		uniqueID:  len(a.users),
		id:        u.ID,
		name:      u.Username,
		reminders: make([]item, 0, initItemBufferCap),
		tasks:     make([]item, 0, initItemBufferCap),
	}
	err := a.db.Exec("INSERT INTO users (id, discord_id, name) VALUES (?, ?, ?);", newUser.uniqueID, newUser.id, newUser.name)
	if err != nil {
		return err
	}
	a.users[u.ID] = newUser
	return nil
}

func (a *app) removeItem(u *discordgo.User, itemName string, kind itemKind) {
	if user, exist := a.users[u.ID]; exist {
		var removed item
		switch kind {
		case itemReminder:
			index := findItemByName(user.reminders, itemName)
			if index == -1 {
				log.Panicf("No item with name %s", itemName)
				return
			}
			removed = user.reminders[index]
			if len(user.reminders) > 1 {
				copy(user.reminders[index-1:], user.reminders[index:])
				user.reminders = user.reminders[:len(user.reminders)-1]
			} else {
				user.reminders = user.reminders[:0]
			}

		case itemTask:
			index := findItemByName(user.tasks, itemName)
			if index == -1 {
				log.Panicf("No item with name %s", itemName)
				return
			}
			removed = user.tasks[index]
			if len(user.tasks) > 1 {
				copy(user.tasks[index-1:], user.tasks[index:])
				user.tasks = user.tasks[:len(user.reminders)-1]
			} else {
				user.tasks = user.tasks[:0]
			}
		}
		err := a.db.Exec("DELETE FROM items WHERE id = ?;", removed.id)
		if err != nil {
			log.Println("DB access failure: ", err)
		}
	}
}

func (a *app) genItemID() int {
	itemID := atomic.SwapInt32(&a.config.ItemCounter, a.config.ItemCounter+1)
	return int(itemID)
}
