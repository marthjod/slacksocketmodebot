package main

import (
	"fmt"
	"os"

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

	// TODO: integrate into struct
	go listen(socketMode, client, logger)

	socketMode.Run()
}

func listen(socketMode *socketmode.Client, webApi *slack.Client, logger zerolog.Logger) {
	for envelope := range socketMode.Events {
		switch envelope.Type {
		case socketmode.EventTypeEventsAPI:
			// Events API:

			// Acknowledge the eventPayload first
			socketMode.Ack(*envelope.Request)

			eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)
			switch eventPayload.Type {
			case slackevents.CallbackEvent:
				switch event := eventPayload.InnerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					logger.Info().Str("user", event.User).Msgf("event text: %v", event.Text)

					if _, _, err := webApi.PostMessage(
						event.Channel,
						slack.MsgOptionText(
							fmt.Sprintf(":wave: <@%v>", event.User),
							false,
						),
					); err != nil {
						logger.Warn().Err(err).Msg("failed to reply")
					}
				default:
					socketMode.Debugf("Skipped: %v", event)
				}
			default:
				socketMode.Debugf("unsupported Events API eventPayload received")
			}

		default:
			socketMode.Debugf("Skipped: %v", envelope.Type)
		}
	}
}
