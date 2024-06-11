package main

import (
	"context"
	"fmt"
	"github.com/ChrisMcKee/hanu"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"log"
	"time"
)

func init() {
	Register("coffee",
		"Reply with the coffee action dialog",
		func(conv hanu.Convo) {
			attachment := slack.Attachment{
				Text:       "I am Coffeebot :robot_face:, and I'm here to help bring you fresh coffee :coffee:",
				Color:      "#3AA3E3",
				CallbackID: Bot.ID + "coffee_order_form",
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
			if _, _, err := Bot.SocketClient.Client.PostMessage(conv.Message().Channel(), options); err != nil {
				fmt.Printf("failed to post message: %s", err)
			}
		})

	RegisterDialogInteraction(hanu.DialogCfg{
		Type:       hanu.Dialog,
		CallbackId: "coffee_order_form",
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
	})
}

// Services the command
var coffeeOrders = make(map[string]map[string]string)

func coffeeRequest(message slack.InteractionCallback, client *socketmode.Client) (slack.Message, bool) {
	if _, ok := coffeeOrders[message.User.ID]["MessageTs"]; !ok {
		coffeeOrders[message.User.ID] = make(map[string]string)
	}
	coffeeOrders[message.User.ID]["MessageTs"] = message.MessageTs
	dialog := makeDialog(message.User.ID)
	if err := client.OpenDialogContext(context.TODO(), message.TriggerID, *dialog); err != nil {
		log.Print("open dialog failed: ", err)
		return slack.Message{}, true
	}
	msg := message.OriginalMessage
	msg.ReplaceOriginal = true
	msg.Timestamp = coffeeOrders[message.User.ID]["order_channel"]
	msg.Text = ":pencil: Taking your order..."
	msg.Attachments = []slack.Attachment{}
	return msg, false
}

func makeDialog(userID string) *slack.Dialog {
	return &slack.Dialog{
		Title:       "Request a coffee",
		SubmitLabel: "Submit",
		CallbackID:  userID + "coffee_order_form",
		Elements: []slack.DialogElement{
			slack.DialogInputSelect{
				DialogInput: slack.DialogInput{
					Label:       "Coffee Type",
					Type:        slack.InputTypeSelect,
					Name:        "mealPreferences",
					Placeholder: "Select a drink",
				},
				Options: []slack.DialogSelectOption{
					{
						Label: "Cappuccino",
						Value: "cappuccino",
					},
					{
						Label: "Latte",
						Value: "latte",
					},
					{
						Label: "Pour Over",
						Value: "pourOver",
					},
					{
						Label: "Cold Brew",
						Value: "coldBrew",
					},
				},
			},
			slack.DialogInput{
				Label:    "Customization orders",
				Type:     slack.InputTypeTextArea,
				Name:     "customizePreference",
				Optional: true,
			},
		},
	}
}
