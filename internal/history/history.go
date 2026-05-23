package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	ID        int       `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	OS        string    `json:"os"`
	CreatedAt time.Time `json:"created_at"`
}

func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sage", "history.json"), nil
}

func Load() ([]Entry, error) {
	p, err := historyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func Append(entry Entry) error {
	entries, err := Load()
	if err != nil {
		entries = []Entry{}
	}
	entry.ID = len(entries) + 1
	entry.CreatedAt = time.Now()
	entries = append(entries, entry)

	p, err := historyPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func Clear() error {
	p, err := historyPath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte("[]"), 0644)
}
