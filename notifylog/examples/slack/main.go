package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/thienhaole92/uframework/notifylog"
	"github.com/thienhaole92/uframework/notifylog/notifier"
)

func main() {
	slack := notifier.NewSlackNotifier(
		zerolog.InfoLevel,
		os.Getenv("SLACK_CHANNEL"),
		slack.New(os.Getenv("SLACK_TOKEN")),
	)
	log := notifylog.New("test", notifylog.JSON, slack)

	log.Info().Str("foo", "bar").Msg("Hello world ddd")
}
