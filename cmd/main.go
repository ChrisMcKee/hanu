package main

import (
	"context"
	"fmt"
	"github.com/ChrisMcKee/hanu"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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

		bot.RegisterDialogInteraction(hanu.DialogCfg{
			Dialog: func(b *hanu.Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
				msg, done := coffeeRequest(cb, client)
				if done {
					client.Ack(*evt.Request, msg)
					return nil
				}
				return nil
			},
			SubmissionHandler: func(b *hanu.Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
				client.Debugf("Order received: %+v\n", cb.Submission)

				go func() {
					time.Sleep(time.Second * 5)

					attachment := slack.Attachment{
						Text:       ":white_check_mark: Order received!",
						CallbackID: b.ID + "coffee_order_form",
					}
					options := slack.MsgOptionAttachments(attachment)
					if _, _, err := client.PostMessage(cb.Channel.ID, options); err != nil {
						log.Print("[ERROR] Failed to post message")
					}
					return
				}()

				client.Ack(*evt.Request)
				return nil
			},
			CallbackId: "coffee_order_form",
		})

		bot.RegisterSlashCommand("/modaltest", func(evt *socketmode.Event, client *socketmode.Client) {
			modalRequest := generateModalRequest()
			eventsAPIEvent, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				fmt.Printf("Ignored %+v\n", evt)
				return
			}
			_, err := client.OpenView(eventsAPIEvent.TriggerID, modalRequest)
			if err != nil {
				fmt.Printf("failed opening view: %v", err)
			}
			client.Ack(*evt.Request)
		})

		bot.RegisterInteraction(slack.InteractionTypeViewSubmission, func(evt *socketmode.Event, client *socketmode.Client) {
			event, ok := evt.Data.(slack.InteractionCallback)
			if !ok {
				fmt.Printf("Ignored %+v\n", evt)
				return
			}

			updateModal := updateModal()
			_, err := client.UpdateView(updateModal, "", event.View.Hash, event.View.ID)
			// Wait for a few seconds to see result this code is necessary due to slack server modal is going to be closed after the update
			time.Sleep(time.Second * 2)
			if err != nil {
				log.Printf("Error updating view: %s", err)
				return
			}

			client.Debugf("button clicked!")
			client.Ack(*evt.Request)
		})

		ctx, cancel := context.WithCancel(context.Background())

		defer cancel() //cancel when we are finished being a bot

		cmdList := List()
		for i := 0; i < len(cmdList); i++ {
			bot.Register(cmdList[i])
		}

		bot.Listen(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
}
