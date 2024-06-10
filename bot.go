package hanu

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
)

// Bot is the main object
type Bot struct {
	SocketClient      *socketmode.Client
	ID                string
	Commands          []CommandInterface
	ReplyOnly         bool
	CmdPrefix         string
	socketmodeHandler *socketmode.SocketmodeHandler
	unknownCmdHandler Handler
	listenerEnabled   bool
}

// New creates a new bot
func New(token string, appToken string) (*Bot, error) {
	api := slack.New(
		token,
		slack.OptionDebug(false),
		slack.OptionAppLevelToken(appToken),
	)
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(false),
	)

	r, e := api.AuthTest()
	if e != nil {
		return nil, e
	}

	bot := &Bot{
		SocketClient:      socketClient,
		ID:                r.UserID,
		socketmodeHandler: socketmode.NewSocketmodeHandler(socketClient),
	}

	return bot, nil
}

// NewDebug New creates a new bot with Debug
func NewDebug(token string, appToken string) (*Bot, error) {
	api := slack.New(
		token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	r, e := api.AuthTest()
	if e != nil {
		return nil, e
	}

	bot := &Bot{
		SocketClient:      socketClient,
		ID:                r.UserID,
		socketmodeHandler: socketmode.NewSocketmodeHandler(socketClient),
	}

	return bot, nil
}

// SetCommandPrefix will set thing that must be prefixed to the command,
// there is no prefix by default but one could set it to "!" for instance
func (b *Bot) SetCommandPrefix(pfx string) *Bot {
	b.CmdPrefix = pfx
	return b
}

// SetReplyOnly will make the bot only respond to messages it is mentioned in
func (b *Bot) SetReplyOnly(ro bool) *Bot {
	b.ReplyOnly = ro
	return b
}

// Process incoming message
func (b *Bot) process(msg Message) {
	// Strip @BotName from public message
	msg.SetText(msg.StripMention(b.ID))
	// Strip Slack's link markup
	msg.SetText(msg.StripLinkMarkup())

	// Only send auto-generated help command list if directly mentioned
	if msg.IsRelevantFor(b.ID) && msg.IsHelpRequest() {
		b.sendHelp(msg)
		return
	}

	// if bot can only reply, ensure we were mentioned
	if b.ReplyOnly && !msg.IsRelevantFor(b.ID) {
		return
	}

	handled := b.searchCommand(msg)
	if !handled && b.ReplyOnly {
		if b.unknownCmdHandler != nil {
			b.unknownCmdHandler(NewConversation(dummyMatch{}, msg, b))
		}
	}
}

// Search for a command matching the message
func (b *Bot) searchCommand(msg Message) bool {
	var cmd CommandInterface

	for i := 0; i < len(b.Commands); i++ {
		cmd = b.Commands[i]

		match, err := cmd.Get().Match(msg.Text())
		if err == nil {
			cmd.Handle(NewConversation(match, msg, b))
			return true
		}
	}

	return false
}

// Channel will return a channel that the bot can talk in
func (b *Bot) Channel(id string) Channel {
	return Channel{b, id}
}

// Say will cause the bot to say something in the specified channel
func (b *Bot) Say(channel, msg string, a ...interface{}) {
	b.send(Message{ChannelID: channel, Message: fmt.Sprintf(msg, a...)})
}

func (b *Bot) send(msg MessageInterface) {
	_, _, err := b.SocketClient.PostMessage(
		msg.Channel(),
		slack.MsgOptionText(msg.Text(), false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

// BuildHelpText will build the help text
func (b *Bot) BuildHelpText() string {
	var cmd CommandInterface
	help := "The available commands are:\n\n"

	for i := 0; i < len(b.Commands); i++ {
		cmd = b.Commands[i]

		help = help + "`" + b.CmdPrefix + cmd.Get().Text() + "`"
		if cmd.Description() != "" {
			help = help + " *–* " + cmd.Description()
		}

		help = help + "\n"
	}

	return help
}

// sendHelp will send help to the channel and user in the given message
func (b *Bot) sendHelp(msg MessageInterface) {
	help := b.BuildHelpText()

	if !msg.IsDirectMessage() {
		help = "<@" + msg.User() + ">: " + help
	}

	b.Say(msg.Channel(), help)
}

func middlewareEventsAPIWithBot(b *Bot) socketmode.SocketmodeHandlerFunc {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		middlewareEventsAPI(evt, client, b)
	}
}

// Listen for message on socket
func (b *Bot) Listen(ctx context.Context) {
	b.socketmodeHandler.Handle(socketmode.EventTypeConnecting, middlewareConnecting)
	b.socketmodeHandler.Handle(socketmode.EventTypeConnectionError, middlewareConnectionError)
	b.socketmodeHandler.Handle(socketmode.EventTypeConnected, middlewareConnected)

	// Handle all EventsAPI
	b.socketmodeHandler.Handle(socketmode.EventTypeEventsAPI, middlewareEventsAPIWithBot(b))

	// Handle a specific event from EventsAPI
	b.socketmodeHandler.HandleEvents(slackevents.AppMention, middlewareAppMentionEvent)

	b.listenerEnabled = true
	b.socketmodeHandler.RunEventLoopContext(ctx)
}

func middlewareEventsAPI(evt *socketmode.Event, client *socketmode.Client, b *Bot) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		client.Debugf("Ignored %+v\n", evt)
		return
	}

	client.Ack(*evt.Request)

	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			go b.process(NewMessage(ev))
		case *slackevents.AppMentionEvent:
			go b.process(NewMentionMessage(ev))
		}
	default:
		client.Debugf("unhandled Events API event received")
	}
}

func middlewareAppMentionEvent(evt *socketmode.Event, client *socketmode.Client) {

	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		client.Debugf("Ignored %+v\n", evt)
		return
	}

	client.Ack(*evt.Request)

	ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		client.Debugf("Ignored %+v\n", ev)
		return
	}

	fmt.Printf("We have been mentioned in %v\n", ev.Channel)
	_, _, err := client.Client.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

func middlewareConnecting(evt *socketmode.Event, client *socketmode.Client) {
	client.Debugf("Connecting to Slack with Socket Mode...")
}

func middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	client.Debugf("Connection failed. Retrying later...")
}

func middlewareConnected(evt *socketmode.Event, client *socketmode.Client) {
	client.Debugf("Connected to Slack with Socket Mode.")
}

// Command adds a new command with custom handler
func (b *Bot) Command(cmd string, handler Handler) {
	b.Commands = append(b.Commands, NewCommand(b.CmdPrefix+cmd, "", handler))
}

// UnknownCommand will be called when the user calls a command that is unknown,
// but it will only work when the bot is in reply only mode
func (b *Bot) UnknownCommand(h Handler) {
	b.unknownCmdHandler = h
}

// Register registers a Command
func (b *Bot) Register(cmd CommandInterface) {
	b.Commands = append(b.Commands, cmd)
}

func (b *Bot) RegisterSlashCommand(cmd string, handler func(evt *socketmode.Event, client *socketmode.Client)) {
	if b.listenerEnabled {
		log.Fatal("RegisterSlashCommand must be called before Listen")
	}
	b.socketmodeHandler.HandleSlashCommand(cmd, handler)
}

func (b *Bot) RegisterInteraction(et slack.InteractionType, handler func(evt *socketmode.Event, client *socketmode.Client)) {
	if b.listenerEnabled {
		log.Fatal("RegisterSlashCommand must be called before Listen")
	}
	b.socketmodeHandler.HandleInteraction(et, handler)
}

func (b *Bot) RegisterEventHandler(et slackevents.EventsAPIType, handler func(evt *socketmode.Event, client *socketmode.Client)) {
	if b.listenerEnabled {
		log.Fatal("RegisterSlashCommand must be called before Listen")
	}
	if et == slackevents.AppMention {
		log.Fatal("AppMention event type is reserved for Bot")
		return
	}
	b.socketmodeHandler.HandleEvents(et, handler)
}
