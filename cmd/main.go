package main

import (
	"context"
	"fmt"
	"github.com/ChrisMcKee/hanu"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
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

	bot.Register(hanu.NewCommand("coffee",
		"Reply with the coffee action dialog",
		func(conv hanu.Convo) {
			attachment := slack.Attachment{
				Text:       "I am Coffeebot :robot_face:, and I'm here to help bring you fresh coffee :coffee:",
				Color:      "#3AA3E3",
				CallbackID: bot.ID + "coffee_order_form",
				Actions: []slack.AttachmentAction{
					{
						Name:  "coffee_order",
						Text:  ":coffee: Order Coffee",
						Type:  "button",
						Value: "coffee_order",
					},
				},
			}

			options := slack.MsgOptionAttachments(attachment)

			if _, _, err := bot.SocketClient.Client.PostMessage(conv.Message().Channel(), options); err != nil {
				fmt.Printf("failed to post message: %s", err)
			}
		}))

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
	//
	//// You can register the same interaction type multiple times with separate filters.
	//bot.RegisterInteraction(slack.InteractionTypeInteractionMessage, func(evt *socketmode.Event, client *socketmode.Client) {
	//	callback, ok := evt.Data.(slack.InteractionCallback)
	//	if !ok {
	//		fmt.Printf("Ignored %+v\n", evt)
	//		return
	//	}
	//
	//	//var payload interface{}
	//
	//	switch callback.CallbackID {
	//	case bot.ID + "coffee_order_form_x":
	//		msg, done := coffeeRequest(callback, client)
	//		if done {
	//			client.Ack(*evt.Request, msg)
	//			return
	//		}
	//		client.Ack(*evt.Request, msg)
	//		break
	//	}
	//
	//})
	//
	//bot.RegisterInteraction(slack.InteractionTypeInteractionMessage, func(evt *socketmode.Event, client *socketmode.Client) {
	//	callback, ok := evt.Data.(slack.InteractionCallback)
	//	if !ok {
	//		fmt.Printf("Ignored %+v\n", evt)
	//		return
	//	}
	//
	//	//var payload interface{}
	//
	//	switch callback.CallbackID {
	//	case bot.ID + "coffee_order_form":
	//		msg, done := coffeeRequest(callback, client)
	//		if done {
	//			client.Ack(*evt.Request, msg)
	//			return
	//		}
	//		client.Ack(*evt.Request, msg)
	//		break
	//	}
	//
	//})
	//
	//bot.RegisterInteraction(slack.InteractionTypeDialogSubmission, func(evt *socketmode.Event, client *socketmode.Client) {
	//	callback, ok := evt.Data.(slack.InteractionCallback)
	//	if !ok {
	//		fmt.Printf("Ignored %+v\n", evt)
	//		return
	//	}
	//
	//	var payload interface{}
	//	switch callback.CallbackID {
	//	case callback.User.ID + "coffee_order_form":
	//		// Save the order
	//		client.Debugf("Order received: %+v\n", callback.Submission)
	//		client.Ack(*evt.Request, payload)
	//		break
	//	}
	//
	//})

	bot.RegisterDialogInteraction(hanu.DialogCfg{
		Prompt: func(cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
			msg, done := coffeeRequest(cb, client)
			if done {
				client.Ack(*evt.Request, msg)
				return nil
			}
			return nil
		},
		Handler: func(cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
			client.Debugf("Order received: %+v\n", cb.Submission)
			client.Ack(*evt.Request)
			return nil
		},
		CallbackId: "coffee_order_form",
	})

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel() //cancel when we are finished being a bot

	bot.Listen(ctx)

}
