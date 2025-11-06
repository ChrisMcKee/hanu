package hanu

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type DialogueType int64

const (
	Dialog DialogueType = 0
	Modal  DialogueType = 1
)
const (
	dialog string = "dialog"
	modal         = "modal"
)

type DialogEvtHandler func(b *Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error

type DialogCfg struct {
	Type              DialogueType
	Dialog            DialogEvtHandler
	SubmissionHandler DialogEvtHandler
	CallbackId        string
}

// RegisterDialogInteraction registers a dialog interaction.
// Callback ID for launching a dialogue is based on the bot-id and the callbackId;
// for submission it is based on the user-id and the callbackId
func (b *Bot) RegisterDialogInteraction(evtHandlerCfg DialogCfg) {
	if b.listenerEnabled {
		log.Fatal("RegisterSlashCommand must be called before Listen")
	}

	b.RegisterInteraction(slack.InteractionTypeInteractionMessage, func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			fmt.Printf("Ignored %+v\n", evt)
			return
		}

		switch callback.CallbackID {
		case b.ID + evtHandlerCfg.CallbackId:
			err := evtHandlerCfg.Dialog(b, callback, evt, client)
			if err != nil {
				fmt.Printf("Error %+v\n", err)
			}
			client.Ack(*evt.Request)
			break
		}
	})

	b.RegisterInteraction(slack.InteractionTypeDialogSubmission, func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			fmt.Printf("Ignored %+v\n", evt)
			return
		}

		switch callback.CallbackID {
		case callback.User.ID + evtHandlerCfg.CallbackId:
			err := evtHandlerCfg.SubmissionHandler(b, callback, evt, client)
			if err != nil {
				fmt.Printf("Error %+v\n", err)
			}
			client.Ack(*evt.Request)
		}
	})
}

func (b *Bot) RegisterModalInteraction(evtHandlerCfg DialogCfg) {
	if b.listenerEnabled {
		log.Fatal("RegisterSlashCommand must be called before Listen")
	}

	b.RegisterInteraction(slack.InteractionTypeInteractionMessage, func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			fmt.Printf("Ignored %+v\n", evt)
			return
		}

		switch callback.CallbackID {
		case b.ID + evtHandlerCfg.CallbackId:
			err := evtHandlerCfg.Dialog(b, callback, evt, client)
			if err != nil {
				fmt.Printf("Error %+v\n", err)
			}
			client.Ack(*evt.Request)
			break
		}
	})

	b.RegisterInteraction(slack.InteractionTypeViewSubmission, func(evt *socketmode.Event, client *socketmode.Client) {
		callback, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			fmt.Printf("Ignored %+v\n", evt)
			return
		}

		switch callback.View.CallbackID {
		case callback.User.ID + evtHandlerCfg.CallbackId:
			err := evtHandlerCfg.SubmissionHandler(b, callback, evt, client)
			if err != nil {
				fmt.Printf("Error %+v\n", err)
			}
			client.Ack(*evt.Request)
		}
	})
}
