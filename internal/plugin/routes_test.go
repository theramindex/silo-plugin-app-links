package plugin

import (
	"context"
	"strings"
	"testing"

	pluginv1 "github.com/Silo-Server/silo-plugin-sdk/pkg/pluginproto/silo/plugin/v1"
	"github.com/theramindex/silo-plugin-app-links/internal/store"
)

type fakeStore struct {
	links []store.Link
}

func (f *fakeStore) List() ([]store.Link, error) {
	return f.links, nil
}

func (f *fakeStore) Upsert(link store.Link) (store.Link, error) {
	link, err := store.Normalize(link)
	if err != nil {
		return store.Link{}, err
	}
	f.links = append(f.links, link)
	return link, nil
}

func (f *fakeStore) Delete(id string) error {
	next := f.links[:0]
	for _, link := range f.links {
		if link.ID != id {
			next = append(next, link)
		}
	}
	f.links = next
	return nil
}

func (f *fakeStore) Path() string {
	return "/tmp/app-links.json"
}

func TestHandleUserPage(t *testing.T) {
	server := NewHTTPRoutesServer(&fakeStore{})
	response, err := server.Handle(context.Background(), &pluginv1.HandleHTTPRequest{Path: "/app-links", Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetStatusCode() != 200 {
		t.Fatalf("status = %d", response.GetStatusCode())
	}
}

func TestHandleSaveLink(t *testing.T) {
	store := &fakeStore{}
	server := NewHTTPRoutesServer(store)
	response, err := server.Handle(context.Background(), &pluginv1.HandleHTTPRequest{
		Path:   "/app-links/admin/api/links",
		Method: "POST",
		Body:   []byte(`{"name":"Plex","url":"https://plex.example.com","enabled":true}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetStatusCode() != 200 {
		t.Fatalf("status = %d body = %s", response.GetStatusCode(), response.GetBody())
	}
	if len(store.links) != 1 || store.links[0].ID != "plex" {
		t.Fatalf("links = %#v", store.links)
	}
	if store.links[0].OpenMode != "new_tab" {
		t.Fatalf("open mode = %q, want new_tab", store.links[0].OpenMode)
	}
}

func TestAdminPageDefaultsNewLinksToNewTab(t *testing.T) {
	body := adminPageHTML("/tmp/app-links.json")
	if !strings.Contains(body, `<option value="new_tab" selected>New tab</option>`) {
		t.Fatalf("expected admin open mode selector to default to new tab")
	}
}
