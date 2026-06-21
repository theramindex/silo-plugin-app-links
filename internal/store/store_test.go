package store

import (
	"path/filepath"
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
	if link.OpenMode != "iframe" {
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
