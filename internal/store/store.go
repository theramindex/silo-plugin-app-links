package store

import (
	"crypto/sha256"
	"encoding/hex"
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

// ValidationError reports a problem with one form field. Field names match the
// JSON keys of Link so the admin page can highlight the field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"error"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// DuplicateError reports that a new link's name produces the ID of a link that
// already exists, which would otherwise overwrite it.
type DuplicateError struct {
	Name string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("A link named %q already exists. Edit that link or choose a different name.", e.Name)
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
	creating := strings.TrimSpace(link.ID) == ""
	link, err = Normalize(link)
	if err != nil {
		return Link{}, err
	}
	link.UpdatedAt = time.Now().Unix()
	replaced := false
	for index := range links {
		if links[index].ID == link.ID {
			if creating {
				return Link{}, &DuplicateError{Name: links[index].Name}
			}
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
	if link.Name == "" {
		return Link{}, &ValidationError{Field: "name", Message: "Enter a name for the link."}
	}
	if link.ID == "" {
		link.ID = slug(link.Name)
	}
	if link.ID == "" {
		// Names without ASCII letters or digits (for example non-Latin
		// scripts) produce an empty slug; derive a stable ID from the name.
		sum := sha256.Sum256([]byte(link.Name))
		link.ID = "link-" + hex.EncodeToString(sum[:6])
	}
	if !validURL(link.URL) {
		return Link{}, &ValidationError{Field: "url", Message: "Enter a full URL starting with http:// or https://, for example https://app.example.com."}
	}
	if link.IconURL != "" && !validURL(link.IconURL) {
		return Link{}, &ValidationError{Field: "iconUrl", Message: "Enter a full icon URL starting with http:// or https://, or leave it empty."}
	}
	if link.OpenMode == "" {
		link.OpenMode = "new_tab"
	}
	if link.OpenMode != "iframe" && link.OpenMode != "new_tab" {
		return Link{}, &ValidationError{Field: "openMode", Message: "Choose New tab or Inside Silo."}
	}
	return link, nil
}

func validURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
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
