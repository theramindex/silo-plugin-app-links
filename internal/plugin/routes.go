package plugin

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strings"

	pluginv1 "github.com/Silo-Server/silo-plugin-sdk/pkg/pluginproto/silo/plugin/v1"
	"github.com/theramindex/silo-plugin-app-links/internal/store"
)

type Store interface {
	List() ([]store.Link, error)
	Upsert(store.Link) (store.Link, error)
	Delete(string) error
	Path() string
}

type HTTPRoutesServer struct {
	pluginv1.UnimplementedHttpRoutesServer
	store Store
}

func NewHTTPRoutesServer(store Store) *HTTPRoutesServer {
	return &HTTPRoutesServer{store: store}
}

func (s *HTTPRoutesServer) Handle(ctx context.Context, request *pluginv1.HandleHTTPRequest) (*pluginv1.HandleHTTPResponse, error) {
	switch request.GetPath() {
	case "/app-links":
		return htmlResponse(http.StatusOK, userPageHTML()), nil
	case "/app-links/open":
		return s.handleOpen(request)
	case "/app-links/api/links":
		return s.handleList(false)
	case "/app-links/admin":
		return htmlResponse(http.StatusOK, adminPageHTML(s.store.Path())), nil
	case "/app-links/admin/api/links":
		if request.GetMethod() == "POST" {
			return s.handleSave(request)
		}
		return s.handleList(true)
	case "/app-links/admin/api/delete":
		return s.handleDelete(request)
	default:
		return textResponse(http.StatusNotFound, "route not found"), nil
	}
}

func (s *HTTPRoutesServer) handleList(includeDisabled bool) (*pluginv1.HandleHTTPResponse, error) {
	links, err := s.store.List()
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]any{"error": err.Error()}), nil
	}
	if !includeDisabled {
		links = enabledLinks(links)
	}
	return jsonResponse(http.StatusOK, map[string]any{"links": links}), nil
}

func (s *HTTPRoutesServer) handleSave(request *pluginv1.HandleHTTPRequest) (*pluginv1.HandleHTTPResponse, error) {
	var link store.Link
	if err := json.Unmarshal(request.GetBody(), &link); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"error": "invalid json"}), nil
	}
	saved, err := s.store.Upsert(link)
	if err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"error": err.Error()}), nil
	}
	return jsonResponse(http.StatusOK, map[string]any{"link": saved}), nil
}

func (s *HTTPRoutesServer) handleDelete(request *pluginv1.HandleHTTPRequest) (*pluginv1.HandleHTTPResponse, error) {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(request.GetBody(), &payload); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"error": "invalid json"}), nil
	}
	if err := s.store.Delete(payload.ID); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]any{"error": err.Error()}), nil
	}
	return jsonResponse(http.StatusOK, map[string]any{"ok": true}), nil
}

func (s *HTTPRoutesServer) handleOpen(request *pluginv1.HandleHTTPRequest) (*pluginv1.HandleHTTPResponse, error) {
	id := queryValue(request, "id")
	links, err := s.store.List()
	if err != nil {
		return textResponse(http.StatusInternalServerError, err.Error()), nil
	}
	for _, link := range links {
		if link.ID == id && link.Enabled {
			if link.OpenMode == "new_tab" {
				return redirectResponse(link.URL), nil
			}
			return htmlResponse(http.StatusOK, iframePageHTML(link)), nil
		}
	}
	return textResponse(http.StatusNotFound, "app link not found"), nil
}

func queryValue(request *pluginv1.HandleHTTPRequest, key string) string {
	query := request.GetQuery()
	if query == nil {
		return ""
	}
	value, ok := query.AsMap()[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func enabledLinks(links []store.Link) []store.Link {
	out := make([]store.Link, 0, len(links))
	for _, link := range links {
		if link.Enabled {
			out = append(out, link)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return strings.ToLower(out[i].Category) < strings.ToLower(out[j].Category)
		}
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func jsonResponse(status int, payload any) *pluginv1.HandleHTTPResponse {
	data, _ := json.Marshal(payload)
	return &pluginv1.HandleHTTPResponse{
		StatusCode: int32(status),
		Headers:    map[string]string{"content-type": "application/json; charset=utf-8"},
		Body:       data,
	}
}

func htmlResponse(status int, body string) *pluginv1.HandleHTTPResponse {
	return &pluginv1.HandleHTTPResponse{
		StatusCode: int32(status),
		Headers:    map[string]string{"content-type": "text/html; charset=utf-8"},
		Body:       []byte(body),
	}
}

func textResponse(status int, body string) *pluginv1.HandleHTTPResponse {
	return &pluginv1.HandleHTTPResponse{
		StatusCode: int32(status),
		Headers:    map[string]string{"content-type": "text/plain; charset=utf-8"},
		Body:       []byte(body),
	}
}

func redirectResponse(target string) *pluginv1.HandleHTTPResponse {
	return &pluginv1.HandleHTTPResponse{
		StatusCode: http.StatusFound,
		Headers:    map[string]string{"location": target},
	}
}

func esc(value string) string {
	return template.HTMLEscapeString(value)
}
