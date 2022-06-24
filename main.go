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
	db, err := genji.Open("remindme")
	if err != nil {
		log.Panicln(err)
	}

	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Panicln(err)
	}

	session.AddHandler(onReady)
	session.AddHandler(onMessage)

	theApp = &app{
		s:               session,
		remindChannelID: "649758541376127015",
		db:              db,
		shouldClose:     make(chan bool),
		users:           make(map[string]*user),
	}
	theApp.init()
	err = theApp.s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer theApp.s.Close()
	defer theApp.db.Close()
	go theApp.run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	theApp.shouldClose <- true
	log.Println("Graceful shutdown")
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	// _, err := s.ChannelMessageSend("649758541376127015", "I am online!")
	// if err != nil {
	// 	fmt.Println(err)
	// }
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
