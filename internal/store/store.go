package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const DefaultDataFile = "/var/lib/continuum/plugins/silo.ramindex.app-links/app-links.json"

type Link struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	OpenMode    string `json:"openMode"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int    `json:"sortOrder,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

type FileStore struct {
	mu   sync.RWMutex
	path string
}

func New(path string) *FileStore {
	if strings.TrimSpace(path) == "" {
		path = os.Getenv("APP_LINKS_DATA_FILE")
	}
	if strings.TrimSpace(path) == "" {
		path = DefaultDataFile
	}
	return &FileStore{path: path}
}

func (s *FileStore) Path() string {
	return s.path
}

func (s *FileStore) List() ([]Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.read()
}

func (s *FileStore) Upsert(link Link) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	links, err := s.read()
	if err != nil {
		return Link{}, err
	}
	link, err = Normalize(link)
	if err != nil {
		return Link{}, err
	}
	link.UpdatedAt = time.Now().Unix()
	replaced := false
	for index := range links {
		if links[index].ID == link.ID {
			links[index] = link
			replaced = true
			break
		}
	}
	if !replaced {
		links = append(links, link)
	}
	sortLinks(links)
	if err := s.write(links); err != nil {
		return Link{}, err
	}
	return link, nil
}

func (s *FileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("missing id")
	}
	links, err := s.read()
	if err != nil {
		return err
	}
	next := links[:0]
	for _, link := range links {
		if link.ID != id {
			next = append(next, link)
		}
	}
	if len(next) == len(links) {
		return nil
	}
	return s.write(next)
}

func (s *FileStore) read() ([]Link, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Link{}, nil
	}
	if err != nil {
		return nil, err
	}
	var links []Link
	if len(strings.TrimSpace(string(data))) == 0 {
		return []Link{}, nil
	}
	if err := json.Unmarshal(data, &links); err != nil {
		return nil, err
	}
	normalized := links[:0]
	for _, link := range links {
		link, err := Normalize(link)
		if err != nil {
			continue
		}
		normalized = append(normalized, link)
	}
	sortLinks(normalized)
	return normalized, nil
}

func (s *FileStore) write(links []Link) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func Normalize(link Link) (Link, error) {
	link.ID = strings.TrimSpace(link.ID)
	link.Name = strings.TrimSpace(link.Name)
	link.URL = strings.TrimSpace(link.URL)
	link.Description = strings.TrimSpace(link.Description)
	link.Category = strings.TrimSpace(link.Category)
	link.IconURL = strings.TrimSpace(link.IconURL)
	link.OpenMode = strings.TrimSpace(link.OpenMode)
	if link.ID == "" {
		link.ID = slug(link.Name)
	}
	if link.Name == "" {
		return Link{}, errors.New("name is required")
	}
	if link.ID == "" {
		return Link{}, errors.New("id is required")
	}
	if err := validateURL(link.URL); err != nil {
		return Link{}, fmt.Errorf("url: %w", err)
	}
	if link.IconURL != "" {
		if err := validateURL(link.IconURL); err != nil {
			return Link{}, fmt.Errorf("icon url: %w", err)
		}
	}
	if link.OpenMode == "" {
		link.OpenMode = "iframe"
	}
	if link.OpenMode != "iframe" && link.OpenMode != "new_tab" {
		return Link{}, errors.New("open mode must be iframe or new_tab")
	}
	return link, nil
}

func validateURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("must be http or https")
	}
	if parsed.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func sortLinks(links []Link) {
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].SortOrder != links[j].SortOrder {
			return links[i].SortOrder < links[j].SortOrder
		}
		return strings.ToLower(links[i].Name) < strings.ToLower(links[j].Name)
	})
}
