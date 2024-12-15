package main

import (
	"bufio"
	"os"
	"strings"
)

// FileStorage implementation
func (fs *FileStorage) getFilePath(platform string) string {
	if platform == "discord" {
		return fs.discordFile
	}
	return fs.slackFile
}

func (fs *FileStorage) SavePostedStory(platform, storyID string) error {
	file, err := os.OpenFile(fs.getFilePath(platform), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(storyID + "\n"); err != nil {
		return err
	}
	return nil
}

func (fs *FileStorage) IsStoryPosted(platform, storyID string) (bool, error) {
	file, err := os.OpenFile(fs.getFilePath(platform), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == storyID {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// DBStorage implementation
func (db *DBStorage) SavePostedStory(platform, storyID string) error {
	_, err := db.db.Exec("INSERT OR IGNORE INTO posted_stories (story_id, platform) VALUES (?, ?)",
		storyID, platform)
	return err
}

func (db *DBStorage) IsStoryPosted(platform, storyID string) (bool, error) {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM posted_stories WHERE story_id = ? AND platform = ?",
		storyID, platform).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
