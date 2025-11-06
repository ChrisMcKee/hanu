package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ChrisMcKee/hanu"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func init() {
	Register("coffee-modal",
		"Reply with the coffee action modal",
		func(conv hanu.Convo) {
			attachment := slack.Attachment{
				Text:       "I am Coffeebot :robot_face:, and I'm here to help bring you fresh coffee :coffee:",
				Color:      "#3AA3E3",
				CallbackID: Bot.ID + "modal_coffee_order_form",
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
		Type:       hanu.Modal,
		CallbackId: "modal_coffee_order_form",
		Dialog: func(b *hanu.Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
			msg, done := coffeeModalRequest(cb, client)
			if done {
				client.Ack(*evt.Request, msg)
				return nil
			}
			return nil
		},
		SubmissionHandler: func(b *hanu.Bot, cb slack.InteractionCallback, evt *socketmode.Event, client *socketmode.Client) error {
			client.Debugf("Order received: %+v\n", cb.Submission)

			updateModal := updateCoffeeModal()
			_, err := client.UpdateView(updateModal, "", cb.View.Hash, cb.View.ID)
			if err != nil {
				log.Printf("Error updating view: %s", err)
				return nil
			}
			time.Sleep(time.Second * 2)

			client.Ack(*evt.Request)
			return nil
		},
	})
}

// Services the command
var coffeeModalOrders = make(map[string]map[string]string)

func coffeeModalRequest(message slack.InteractionCallback, client *socketmode.Client) (slack.Message, bool) {
	if _, ok := coffeeModalOrders[message.User.ID]["MessageTs"]; !ok {
		coffeeModalOrders[message.User.ID] = make(map[string]string)
	}
	coffeeModalOrders[message.User.ID]["MessageTs"] = message.MessageTs
	coffeeModal := makeCoffeeModal(message.User.ID)
	if _, err := client.OpenView(message.TriggerID, coffeeModal); err != nil {
		log.Print("open modal failed: ", err)
		return slack.Message{}, true
	}
	msg := message.OriginalMessage
	msg.ReplaceOriginal = true
	msg.Timestamp = coffeeModalOrders[message.User.ID]["order_channel"]
	msg.Text = ":pencil: Taking your order..."
	msg.Attachments = []slack.Attachment{}
	return msg, false
}

func makeCoffeeModal(userID string) slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "Coffee Modal", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	// Coffee Select
	coffeeOptions := createOptionBlockObjects([]string{"Cappuccino", "Latte", "Pour Over", "Cold Brew"}, false)
	coffeeOptionsText := slack.NewTextBlockObject(slack.PlainTextType, "Invitee from static list", false, false)
	coffeeOption := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "coffee-options", coffeeOptions...)
	coffeeOptionsBlock := slack.NewInputBlock("invitee", coffeeOptionsText, nil, coffeeOption)

	customisationText := slack.NewTextBlockObject("plain_text", "Customisation", false, false)
	customisationHint := slack.NewTextBlockObject("plain_text", "How would you like to customise your drink", false, false)
	customisationPlaceHolder := slack.NewTextBlockObject("plain_text", "Enter your options", false, false)
	customisationElement := slack.NewPlainTextInputBlockElement(customisationPlaceHolder, "customisation")
	// Notice that blockID is a unique identifier for a block
	customisationBlock := slack.NewInputBlock("First Name", customisationText, customisationHint, customisationElement)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			coffeeOptionsBlock,
			customisationBlock,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.CallbackID = userID + "modal_coffee_order_form"
	return modalRequest
}

func updateCoffeeModal() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "My App", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Order Taken!", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	return modalRequest
}

func createOptionBlockObjects(options []string, users bool) []*slack.OptionBlockObject {
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))
	var text string
	for _, o := range options {
		if users {
			text = fmt.Sprintf("<@%s>", o)
		} else {
			text = o
		}
		optionText := slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(o, optionText, nil))
	}
	return optionBlockObjects
}
