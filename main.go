package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	hnTopStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	hnItemURL       = "https://hacker-news.firebaseio.com/v0/item/%d.json"
)

type Story struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Config struct {
	DiscordWebhookURL     string
	SlackWebhookURL       string
	SchedulePeriod        int
	FetchTopStoriesAmount int
	StorageType           string
	DBPath                string
}

func loadConfig() Config {
	return Config{
		DiscordWebhookURL:     os.Getenv("DISCORD_WEBHOOK_URL"),
		SlackWebhookURL:       os.Getenv("SLACK_WEBHOOK_URL"),
		SchedulePeriod:        getEnvAsInt("SCHEDULE_PERIOD", 60),
		FetchTopStoriesAmount: getEnvAsInt("FETCH_TOP_STORIES_AMOUNT", 5),
		StorageType:           getEnvWithDefault("STORAGE", "local"),
		DBPath:                "/app/data/posted_stories.db",
	}
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvWithDefault(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

type Storage interface {
	SavePostedStory(platform string, storyID string) error
	IsStoryPosted(platform string, storyID string) (bool, error)
}

type FileStorage struct {
	discordFile string
	slackFile   string
}

type DBStorage struct {
	db *sql.DB
}

func NewStorage(config Config) (Storage, error) {
	if config.StorageType == "database" {
		db, err := sql.Open("sqlite3", config.DBPath)
		if err != nil {
			return nil, err
		}

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posted_stories (
			story_id TEXT,
			platform TEXT,
			PRIMARY KEY (story_id, platform)
		)`)
		if err != nil {
			return nil, err
		}

		return &DBStorage{db: db}, nil
	}

	return &FileStorage{
		discordFile: "discord_posted_stories.txt",
		slackFile:   "slack_posted_stories.txt",
	}, nil
}

func fetchTopStories() ([]int, error) {
	resp, err := http.Get(hnTopStoriesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stories []int
	if err := json.NewDecoder(resp.Body).Decode(&stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func fetchStoryDetails(id int) (*Story, error) {
	resp, err := http.Get(fmt.Sprintf(hnItemURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var story Story
	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return nil, err
	}
	return &story, nil
}

func postToWebhook(platform, webhookURL, message string) error {
	var payload map[string]interface{}

	if platform == "discord" {
		payload = map[string]interface{}{
			"content":  message,
			"username": "Hacker News Bot",
		}
	} else {
		payload = map[string]interface{}{
			"text": message,
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var resp *http.Response
	for attempts := 0; attempts < 3; attempts++ {
		resp, err = http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("Received 429 Too Many Requests, retrying in 1 second...")
			time.Sleep(time.Minute)
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
		}

		log.Printf("Posting to webhook URL: %s with payload: %s", webhookURL, string(jsonData))
		return nil
	}

	return fmt.Errorf("failed to post to webhook after multiple attempts")
}

func fetchAndPostNews(config Config, storage Storage) error {
	platforms := map[string]string{
		"discord": config.DiscordWebhookURL,
		"slack":   config.SlackWebhookURL,
	}

	topStories, err := fetchTopStories()
	if err != nil {
		return fmt.Errorf("failed to fetch top stories: %v", err)
	}

	for platform, webhookURL := range platforms {
		if webhookURL == "" {
			continue
		}

		newStoriesFound := 0
		for _, storyID := range topStories {
			if newStoriesFound >= config.FetchTopStoriesAmount {
				break
			}

			posted, err := storage.IsStoryPosted(platform, strconv.Itoa(storyID))
			if err != nil {
				log.Printf("Error checking if story %d is posted: %v", storyID, err)
				continue
			}

			if posted {
				continue
			}

			story, err := fetchStoryDetails(storyID)
			if err != nil {
				log.Printf("Error fetching story %d details: %v", storyID, err)
				continue
			}

			if story.URL == "" {
				continue
			}

			var message string
			if platform == "discord" {
				message = fmt.Sprintf("[**%s**](%s)    [(comment)](%s)", story.Title, story.URL, "https://news.ycombinator.com/item?id="+strconv.Itoa(story.ID))
			} else {
				message = fmt.Sprintf("%s\n%s", story.Title, story.URL)
			}

			if err := postToWebhook(platform, webhookURL, message); err != nil {
				log.Printf("Error posting to %s: %v", platform, err)
				return err
			}

			if err := storage.SavePostedStory(platform, strconv.Itoa(storyID)); err != nil {
				log.Printf("Error saving posted story %d: %v", storyID, err)
			}

			newStoriesFound++
		}
	}
	return nil
}

func main() {
	config := loadConfig()
	storage, err := NewStorage(config)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	ticker := time.NewTicker(time.Duration(config.SchedulePeriod) * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	if err := fetchAndPostNews(config, storage); err != nil {
		log.Printf("Error in fetch and post routine: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := fetchAndPostNews(config, storage); err != nil {
				log.Printf("Error in fetch and post routine: %v", err)
			}
		}
	}
}
