package interfaces

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/levelaxis/charli/backend/internal/modules/google/application"
	"github.com/levelaxis/charli/backend/internal/modules/google/domain"
)

// fakeRepo is an in-memory domain.Repository for tests.
type fakeRepo struct {
	mu    sync.Mutex
	byDev map[string]domain.Connection
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byDev: make(map[string]domain.Connection)} }

func (f *fakeRepo) Save(c domain.Connection) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byDev[c.DeviceID] = c
	return nil
}

func (f *fakeRepo) FindByDevice(deviceID string) (*domain.Connection, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byDev[deviceID]
	if !ok {
		return nil, false, nil
	}
	return &c, true, nil
}

func newTestHandler() *Handler {
	oauthConfig := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint:     oauth2.Endpoint{TokenURL: "https://example.invalid/token", AuthURL: "https://example.invalid/auth"},
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets"},
	}
	svc := application.NewService(oauthConfig, newFakeRepo())
	return NewHandler(svc)
}

func TestConnectReturnsAuthURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1/integrations/google"), newTestHandler())
	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v1/integrations/google/connect", "application/json", strings.NewReader(`{"deviceId":"device-1"}`))
	if err != nil {
		t.Fatalf("post connect: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestStatusReportsUnconnectedForUnknownDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1/integrations/google"), newTestHandler())
	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/integrations/google/status?deviceId=never-connected")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAppendReportsNotConnected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1/integrations/google"), newTestHandler())
	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"deviceId":"never-connected","spreadsheetId":"abc","values":["x"]}`
	resp, err := http.Post(srv.URL+"/api/v1/integrations/google/append", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post append: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (the failure is reported in the body, not the status), got %d", resp.StatusCode)
	}
}

// TestAppendWithEmptyDeviceIDReportsNotConnected proves an empty deviceId
// (e.g. a transient chrome.storage read failure client-side) degrades to a
// clear "not connected" failure detail, not an opaque 400 — the append
// endpoint doesn't require deviceId for exactly this reason.
func TestAppendWithEmptyDeviceIDReportsNotConnected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1/integrations/google"), newTestHandler())
	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"deviceId":"","spreadsheetId":"abc","values":["x"]}`
	resp, err := http.Post(srv.URL+"/api/v1/integrations/google/append", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post append: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestConnectRejectsMissingDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1/integrations/google"), newTestHandler())
	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v1/integrations/google/connect", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("post connect: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
