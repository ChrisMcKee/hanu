package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChrisMcKee/hanu"
)

var Start time.Time

func main() {
	go func() {
		bot, err := hanu.NewDebug(os.Getenv("SLACKTOKEN"), os.Getenv("SLACKAPPTOKEN"))
		if err != nil {
			log.Fatal(err)
		}

		Bot = bot

		bot.Register(hanu.NewCommand("uptime",
			"Reply with the uptime",
			func(conv hanu.Convo) {
				conv.Reply("Thanks for asking! I'm running since `%s`", time.Since(Start))
			}))

		bot.Register(hanu.NewCommand("getuser",
			"Get user info",
			func(conv hanu.Convo) {
				myUser, err := bot.GetUserName(conv)
				if err != nil {
					conv.Reply("Error getting user info: %s", err)
					return
				}
				conv.Reply("Thanks for asking! Your username is `%s`", myUser)
			}))

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel() //cancel when we are finished being a bot

		cmdList := List()
		for i := 0; i < len(cmdList); i++ {
			bot.Register(cmdList[i])
		}

		diagList := ListDialogInteractions()
		for i := 0; i < len(diagList); i++ {
			switch diagList[i].Type {
			case hanu.Dialog:
				bot.RegisterDialogInteraction(diagList[i])
			case hanu.Modal:
				bot.RegisterModalInteraction(diagList[i])
			}
		}

		bot.Listen(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
}
