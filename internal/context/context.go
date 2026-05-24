package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Session holds the last Q&A so follow-up questions have context.
type Session struct {
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	OS        string    `json:"os"`
	SavedAt   time.Time `json:"saved_at"`
}

func contextPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sage", "context.json"), nil
}

func Save(s Session) error {
	p, err := contextPath()
	if err != nil {
		return err
	}
	s.SavedAt = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// Load returns the last session, or nil if none exists or it's older than 30 minutes.
func Load() *Session {
	p, err := contextPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	// Expire after 30 minutes
	if time.Since(s.SavedAt) > 30*time.Minute {
		return nil
	}
	return &s
}

// IsFollowUp returns true if the question looks like a follow-up.
func IsFollowUp(q string) bool {
	followUpPhrases := []string{
		"why", "how", "what about", "what if", "can you",
		"and ", "but ", "also ", "fix", "fix it", "fix that",
		"explain", "more", "tell me more", "elaborate",
		"how do i", "how to", "what does", "what is",
	}
	lower := make([]byte, len(q))
	for i, c := range q {
		if c >= 'A' && c <= 'Z' {
			lower[i] = byte(c + 32)
		} else {
			lower[i] = byte(c)
		}
	}
	lq := string(lower)
	for _, phrase := range followUpPhrases {
		if len(lq) > 0 && (lq == phrase || len(lq) < 30 && containsWord(lq, phrase)) {
			return true
		}
	}
	return false
}

func containsWord(s, word string) bool {
	if len(s) >= len(word) && s[:len(word)] == word {
		return true
	}
	idx := 0
	for idx < len(s) {
		pos := indexOf(s[idx:], word)
		if pos == -1 {
			return false
		}
		pos += idx
		if pos == 0 || s[pos-1] == ' ' {
			return true
		}
		idx = pos + 1
	}
	return false
}

func indexOf(s, sub string) int {
	if len(sub) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
