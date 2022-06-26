package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/genjidb/genji"
)

const botToken = "OTg5NjM0NTczMjM3Mzc0OTc2.GDTNFt.7wNZR3_WGMLpI5rUvtB1ZSByCzXEcSslcknitU"

func main() {
	db, err := genji.Open("./data/remindme")
	if err != nil {
		log.Panicln(err)
	}

	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Panicln(err)
	}

	session.AddHandler(onMessage)
	session.AddHandler(onReaction)

	theApp = &app{
		s:               session,
		remindChannelID: "649758541376127015",
		db:              db,
		shouldClose:     make(chan bool),
		users:           make(map[string]*user),
	}
	theApp.init()
	defer theApp.shutdown()

	err = theApp.s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer theApp.s.Close()
	defer theApp.db.Close()
	go theApp.run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	theApp.shouldClose <- true
	log.Println("Graceful shutdown")
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	cmd, err := parseCommand(m.Content)
	if !err.isOK() {
		theApp.handleError(m.ChannelID, err)
		return
	}
	theApp.handleCommand(m.Message, cmd)
}

func onReaction(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}
	if m.Emoji.Name != "☑" {
		return
	}
	msg, _ := s.ChannelMessage(m.ChannelID, m.MessageID)
	if !msg.Author.Bot {
		return
	}
	if len(msg.Embeds) == 1 {
		var kind itemKind
		e := msg.Embeds[0]
		switch e.Title {
		case reminderAlarm:
			kind = itemReminder
		case taskAlarm:
			kind = itemTask
		default:
			return
		}

		var start int
		for i := len(e.Description) - 4; i >= 0; i -= 1 {
			if e.Description[i] == '*' {
				start = i + 1
				break
			}
		}
		itemName := e.Description[start : len(e.Description)-3]

		theApp.removeItem(m.Member.User, itemName, kind)
	}
}
