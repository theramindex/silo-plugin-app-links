package store

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeDefaultsIDAndOpenMode(t *testing.T) {
	link, err := Normalize(Link{Name: "Plex Server", URL: "https://plex.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if link.ID != "plex-server" {
		t.Fatalf("ID = %q", link.ID)
	}
	if link.OpenMode != "new_tab" {
		t.Fatalf("OpenMode = %q", link.OpenMode)
	}
}

func TestFileStoreUpsertAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.json")
	s := New(path)
	saved, err := s.Upsert(Link{Name: "Silo", URL: "https://silo.example.com", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	links, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].ID != saved.ID {
		t.Fatalf("links = %#v", links)
	}
	if err := s.Delete(saved.ID); err != nil {
		t.Fatal(err)
	}
	links, err = s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Fatalf("links after delete = %#v", links)
	}
}

func TestUpsertRejectsDuplicateNameOnCreate(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "links.json"))
	if _, err := s.Upsert(Link{Name: "Plex", URL: "https://plex.example.com"}); err != nil {
		t.Fatal(err)
	}
	_, err := s.Upsert(Link{Name: "plex", URL: "https://other.example.com"})
	var duplicate *DuplicateError
	if !errors.As(err, &duplicate) {
		t.Fatalf("err = %v, want DuplicateError", err)
	}
	links, _ := s.List()
	if len(links) != 1 || links[0].URL != "https://plex.example.com" {
		t.Fatalf("existing link was changed: %#v", links)
	}
}

func TestUpsertUpdatesExistingLinkByID(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "links.json"))
	if _, err := s.Upsert(Link{Name: "Plex", URL: "https://plex.example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Upsert(Link{ID: "plex", Name: "Plex", URL: "https://new.example.com"}); err != nil {
		t.Fatal(err)
	}
	links, _ := s.List()
	if len(links) != 1 || links[0].URL != "https://new.example.com" {
		t.Fatalf("links = %#v", links)
	}
}

func TestNormalizeReturnsFieldErrors(t *testing.T) {
	cases := map[string]Link{
		"name":    {URL: "https://a.example.com"},
		"url":     {Name: "A", URL: "a.example.com"},
		"iconUrl": {Name: "A", URL: "https://a.example.com", IconURL: "ftp://x"},
	}
	for field, link := range cases {
		_, err := Normalize(link)
		var validation *ValidationError
		if !errors.As(err, &validation) || validation.Field != field {
			t.Fatalf("%s: err = %v", field, err)
		}
		if strings.Contains(validation.Message, "parse") {
			t.Fatalf("%s: message leaks parser detail: %q", field, validation.Message)
		}
	}
}

func TestNormalizeGivesNonLatinNamesAnID(t *testing.T) {
	link, err := Normalize(Link{Name: "日本", URL: "https://a.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(link.ID, "link-") {
		t.Fatalf("id = %q", link.ID)
	}
}
