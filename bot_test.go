package hanu

//var Start time.Time
//
//func TestBot(t *testing.T) {
//	bot, err := NewDebug(os.Getenv("SLACKTOKEN"), os.Getenv("SLACKAPPTOKEN"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bot.Register(NewCommand("uptime",
//		"Reply with the uptime",
//		func(conv Convo) {
//			conv.Reply("Thanks for asking! I'm running since `%s`", time.Since(Start))
//		}))
//
//	ctx, cancel := context.WithCancel(context.Background())
//
//	defer cancel() //cancel when we are finished being a bot
//
//	bot.Listen(ctx)
//}
