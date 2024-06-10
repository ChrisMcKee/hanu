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

	bot.Listen(ctx)

}

func generateModalRequest() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "My App", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Please enter your name", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	firstNameText := slack.NewTextBlockObject("plain_text", "First Name", false, false)
	firstNameHint := slack.NewTextBlockObject("plain_text", "First Name Hint", false, false)
	firstNamePlaceholder := slack.NewTextBlockObject("plain_text", "Enter your first name", false, false)
	firstNameElement := slack.NewPlainTextInputBlockElement(firstNamePlaceholder, "firstName")
	// Notice that blockID is a unique identifier for a block
	firstName := slack.NewInputBlock("First Name", firstNameText, firstNameHint, firstNameElement)

	lastNameText := slack.NewTextBlockObject("plain_text", "Last Name", false, false)
	lastNameHint := slack.NewTextBlockObject("plain_text", "Last Name Hint", false, false)
	lastNamePlaceholder := slack.NewTextBlockObject("plain_text", "Enter your first name", false, false)
	lastNameElement := slack.NewPlainTextInputBlockElement(lastNamePlaceholder, "lastName")
	lastName := slack.NewInputBlock("Last Name", lastNameText, lastNameHint, lastNameElement)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			firstName,
			lastName,
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

func updateModal() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "My App", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Modal updated!", false, false)
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
