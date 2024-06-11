package hanu

import (
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"log"
)

type DialogEvtHandler func(b *Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error

type DialogCfg struct {
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
