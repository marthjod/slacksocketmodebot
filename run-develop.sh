#!/bin/bash

set -o nounset
set -o errexit
set -o errtrace
set -o pipefail

slack_app_token="$(cat slack-app-token)"
slack_bot_token="$(cat slack-bot-token)"

go build -o cmd

export SLACK_APP_TOKEN="${slack_app_token}"
export SLACK_BOT_TOKEN="${slack_bot_token}"
export SLACK_CHANNEL="chatops-dev"
export LOGLEVEL="debug"

./cmd | jq -R 'fromjson? | .'
