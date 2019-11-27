# hanu `<forked>` - Go for Slack Bots!

- [![MIT License](https://badgen.now.sh/badge/License/MIT/blue)](LICENSE.md)

The `Go` framework **hanu** is your best friend to create [Slack](https://slackhq.com) bots! **hanu** uses [allot](https://github.com/ChrisMcKee/allot) for easy command and request parsing (e.g. `whisper <word>`) and runs fine as a [Heroku worker](https://devcenter.heroku.com/articles/background-jobs-queueing). All you need is a [Slack API token](https://api.slack.com/bot-users) and you can create your first bot within seconds! Just have a look at the [hanu-example](https://github.com/sbstjn/hanu-example) bot or [read my tutorial](https://sbstjn.com/host-golang-slackbot-on-heroku-with-hanu.html) …

### Features

- Respond to **mentions**
- Respond to **direct messages**
- Auto-Generated command list for `help`
- Works fine as a **worker** on Heroku


## V2 Usage

To use the package import:

    import "github.com/ChrisMcKee/hanu"

It is very similar to the above, but there are a few extra things.  You can set the
command prefix, if you like using those:

```
slack.SetCommandPrefix("!")
slack.SetReplyOnly(false)
```

This will make it so you have to type:

```
!whisper I love turtles
```

For the command to be recognised.  Setting the bot to not reply only means it will listen to
all messages in an attempt to find a command (except help will only be printed when bot is mentioned).

Also, the `ConversationInterface` was changed to just `Convo` to save your wrists:

```
	slack.Command("whisper <word>", func(conv hanu.Convo) {
		str, _ := conv.String("word")
		conv.Reply(strings.ToLower(str))
	})
```

The bot can also now talk arbitrarily:

```
slack.Say("UGHXISDF324", "I like %s", "turtles")

devops := slack.Channel("UGHXISDF324")
devops.Say("Host called %s is not responding to pings", "bobsburgers01")
```

You can print the help message whenever you want:

```
slack.Say("UGHXISDF324", bot.BuildHelpText())
```

And there is an unknown command handler, but it only works when in reply only mode:

```
slack.SetReplyOnly(true).UnknownCommand(func(c hanu.Convo) {
	c.Reply(slack.BuildHelpText())
})
```

## Dependencies

- [github.com/ChrisMcKee/allot](https://github.com/ChrisMcKee/allot) for parsing `cmd <param1:string> <param2:integer>` strings
- [golang.org/x/net/websocket](http://golang.org/x/net/websocket) for websocket communication with Slack
- [github.com/nlopes/slack](http://github.com/nlopes/slack) for real time communication with Slack

## Credits

- [Host Go Slackbot on Heroku](https://sbstjn.com/host-golang-slackbot-on-heroku-with-hanu.html)
- [OpsDash article about Slack Bot](https://www.opsdash.com/blog/slack-bot-in-golang.html)
- [A Simple Slack Bot in Go - The Bot](ttps://dev.to/shindakun/a-simple-slack-bot-in-go---the-bot-4olg)


## Forked (along with allot) from  

- [![Read Tutorial](https://badgen.now.sh/badge/Read/Tutorial/orange)](https://sbstjn.com/host-golang-slackbot-on-heroku-with-hanu.html)
- [![Code Example](https://badgen.now.sh/badge/Code/Example/cyan)](https://github.com/sbstjn/hanu-example)
