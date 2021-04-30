package listener

import (
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Listener will listen for app_mention events.
type Listener struct {
	socketModeClient *socketmode.Client
	webAPIClient     *slack.Client
	logger           zerolog.Logger
}

// New returns a ready-to-use Listener.
func New(socketModeClient *socketmode.Client, webAPIClient *slack.Client, logger zerolog.Logger) *Listener {
	return &Listener{
		socketModeClient: socketModeClient,
		webAPIClient:     webAPIClient,
		logger:           logger,
	}
}

// Listen is supposed to be called as a goroutine.
func (li *Listener) Listen(errChan chan<- error, outChan chan<- slackevents.AppMentionEvent) {
	for envelope := range li.socketModeClient.Events {
		switch envelope.Type {
		case socketmode.EventTypeEventsAPI:

			// Acknowledge the eventPayload first
			li.socketModeClient.Ack(*envelope.Request)

			eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)
			switch eventPayload.Type {
			case slackevents.CallbackEvent:
				switch event := eventPayload.InnerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					li.logger.Debug().Str("user", event.User).Str("text", event.Text).Msg("received message")
					// if strings.Contains(event.Text, "error") {
					// 	errChan <- errors.New(event.Text)
					// 	continue
					// }
					outChan <- *event
				default:
					li.socketModeClient.Debugf("skipped: event %v", event)
				}
			default:
				li.socketModeClient.Debugf("unsupported Events API eventPayload received")
			}

		default:
			li.socketModeClient.Debugf("skipped: envelope type %q", envelope.Type)
		}
	}
}
