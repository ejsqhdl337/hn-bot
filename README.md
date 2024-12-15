# Hacker News Bot (Go Version)

A robust Go application that fetches top stories from Hacker News and posts them to Discord and Slack channels via webhooks. The bot features persistent storage options and is containerized for easy deployment.

## Features

- Fetches and posts top Hacker News stories to Discord and Slack
- Configurable posting frequency and number of stories
- Multiple storage options (local file or SQLite database)
- Docker support
- Resilient error handling and continuous operation
- Concurrent webhook posting

## Prerequisites

- Go 1.23 or higher (for local development)
- Docker (for containerized deployment)
- Discord and/or Slack webhook URLs

## Configuration

The application is configured via environment variables:

|
Variable
|
Description
|
Default
|
|

---

## |

## |

|
|
`DISCORD_WEBHOOK_URL`
|
Discord channel webhook URL
|
""
|
|
`SLACK_WEBHOOK_URL`
|
Slack channel webhook URL
|
""
|
|
`SCHEDULE_PERIOD`
|
Posting frequency in minutes
|
60
|
|
`FETCH_TOP_STORIES_AMOUNT`
|
Number of stories to fetch per interval
|
5
|
|
`STORAGE`
|
Storage type ("local" or "database")
|
"local"
|

## Running Locally

1. Clone the repository:

```bash
git clone https://github.com/yourusername/hn-bot
cd hn-bot
```

2. Install dependencies:

```bash
go mod download
```

3. Set environment variables:

```bash
export DISCORD_WEBHOOK_URL="your_discord_webhook_url"
export SLACK_WEBHOOK_URL="your_slack_webhook_url"
export SCHEDULE_PERIOD=60
export FETCH_TOP_STORIES_AMOUNT=5
export STORAGE=local
```

4. Run the application:

```bash
go run .
```

## Docker Deployment

1. Build the Docker image:

```bash
docker build -t hn-bot .
```

2. Run the container:

```bash
docker run -d \
  -e DISCORD_WEBHOOK_URL="your_discord_webhook_url" \
  -e SLACK_WEBHOOK_URL="your_slack_webhook_url" \
  -e SCHEDULE_PERIOD=60 \
  -e FETCH_TOP_STORIES_AMOUNT=5 \
  -e STORAGE=local \
  --name hn-bot \
  hn-bot
```

## Error Handling

The application includes comprehensive error handling:

- Continues running even if API requests fail
- Logs errors without stopping execution
- Retries failed operations where appropriate
- Maintains separate error handling for different platforms
