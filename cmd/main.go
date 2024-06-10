package main

import (
	"context"
	"github.com/ChrisMcKee/hanu"
	"log"
	"os"
	"time"
)

var Start time.Time

func main() {
	bot, err := hanu.NewDebug(os.Getenv("SLACKTOKEN"), os.Getenv("SLACKAPPTOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Register(hanu.NewCommand("uptime",
		"Reply with the uptime",
		func(conv hanu.Convo) {
			conv.Reply("Thanks for asking! I'm running since `%s`", time.Since(Start))
		}))

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel() //cancel when we are finished being a bot

	bot.Listen(ctx)

}
