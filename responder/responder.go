package responder

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Responder struct {
	webAPIClient *slack.Client
	logger       zerolog.Logger
}

func New(webAPIClient *slack.Client, logger zerolog.Logger) *Responder {
	return &Responder{
		webAPIClient: webAPIClient,
		logger:       logger,
	}
}

func (r *Responder) Respond(errChan chan<- error, inChan chan slackevents.AppMentionEvent) {
	for event := range inChan {
		if _, _, err := r.webAPIClient.PostMessage(
			event.Channel,
			slack.MsgOptionText(
				fmt.Sprintf(":wave: <@%v>", event.User),
				false,
			),
		); err != nil {
			r.logger.Warn().Err(err).Msg("failed to reply")
			errChan <- errors.Wrap(err, "failed to reply")
		}
	}
}
