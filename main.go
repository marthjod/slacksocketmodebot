package main

import (
	"os"

	"github.com/marthjod/slacksocketmodebot/listener"
	"github.com/marthjod/slacksocketmodebot/responder"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type wrapLogger struct {
	zerolog.Logger
}

func NewLogger(lo zerolog.Logger) *wrapLogger {
	return &wrapLogger{
		Logger: lo,
	}
}

// see https://golang.org/pkg/log/#Output
func (lo *wrapLogger) Output(_ int, s string) error {
	lo.Logger.Info().Msg(s)
	return nil
}

func main() {
	logger := zerolog.New(os.Stdout)

	client := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		// slack.OptionDebug(true),
		slack.OptionLog(NewLogger(logger.With().Str("component", "api").Logger())),
	)
	socketMode := socketmode.New(
		client,
		// socketmode.OptionDebug(true),
		socketmode.OptionLog(NewLogger(logger.With().Str("component", "socketmode").Logger())),
	)
	if _, authTestErr := client.AuthTest(); authTestErr != nil {
		logger.Fatal().Err(authTestErr).Msg("SLACK_BOT_TOKEN is invalid")
	}

	listener := listener.New(socketMode, client, logger)
	responder := responder.New(client, logger)

	errChan := make(chan error)
	eventChan := make(chan slackevents.AppMentionEvent)

	logger.Debug().Msg("starting event listener")
	go listener.Listen(errChan, eventChan)
	logger.Debug().Msg("starting responder")
	go responder.Respond(errChan, eventChan)

	logger.Debug().Msg("listening for errors")
	go func() {
		for err := range errChan {
			logger.Error().Err(err).Msg("received error")
		}
	}()

	logger.Debug().Msg("starting socketmode")
	socketMode.Run()
}
