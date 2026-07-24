package application

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/levelaxis/charli/backend/internal/modules/google/domain"
)

// fakeRepo is an in-memory domain.Repository for tests.
type fakeRepo struct {
	mu      sync.Mutex
	byDev   map[string]domain.Connection
	saves   int
	saveErr error
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byDev: make(map[string]domain.Connection)} }

func (f *fakeRepo) Save(c domain.Connection) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saves++
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

// fakeTokenServer mimics Google's OAuth token endpoint.
func fakeTokenServer(t *testing.T, accessToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  accessToken,
			"refresh_token": "refreshed-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
}

func testOAuthConfig(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL, AuthURL: "https://accounts.google.com/o/oauth2/auth"},
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets"},
	}
}

func TestUnavailableWithoutConfig(t *testing.T) {
	svc := NewService(nil, newFakeRepo())
	if svc.Available() {
		t.Fatal("expected unavailable without an oauth config")
	}
	if _, err := svc.BeginConnect("device-1"); err == nil {
		t.Fatal("expected BeginConnect to fail when unavailable")
	}
	if svc.IsConnected("device-1") {
		t.Fatal("expected IsConnected to be false when unavailable")
	}
}

func TestUnavailableWithoutRepo(t *testing.T) {
	svc := NewService(testOAuthConfig("https://example.invalid/token"), nil)
	if svc.Available() {
		t.Fatal("expected unavailable without a repository")
	}
}

func TestBeginConnectGeneratesAuthURLWithState(t *testing.T) {
	svc := NewService(testOAuthConfig("https://example.invalid/token"), newFakeRepo())

	authURL, err := svc.BeginConnect("device-1")
	if err != nil {
		t.Fatalf("BeginConnect: %v", err)
	}
	if !strings.Contains(authURL, "state=") {
		t.Fatalf("expected a state param in the auth URL, got %q", authURL)
	}
}

func TestCompleteConnectExchangesCodeAndPersists(t *testing.T) {
	tokenServer := fakeTokenServer(t, "access-token-1")
	defer tokenServer.Close()

	repo := newFakeRepo()
	svc := NewService(testOAuthConfig(tokenServer.URL), repo)

	authURL, err := svc.BeginConnect("device-1")
	if err != nil {
		t.Fatalf("BeginConnect: %v", err)
	}
	state := stateFromAuthURL(t, authURL)

	if err := svc.CompleteConnect(context.Background(), "auth-code", state); err != nil {
		t.Fatalf("CompleteConnect: %v", err)
	}

	conn, found, err := repo.FindByDevice("device-1")
	if err != nil || !found {
		t.Fatalf("expected a persisted connection: found=%v err=%v", found, err)
	}
	if conn.AccessToken != "access-token-1" {
		t.Errorf("expected the exchanged access token to be persisted, got %q", conn.AccessToken)
	}
	if !svc.IsConnected("device-1") {
		t.Error("expected IsConnected to be true after a completed connect")
	}
}

func TestCompleteConnectRejectsUnknownState(t *testing.T) {
	svc := NewService(testOAuthConfig("https://example.invalid/token"), newFakeRepo())
	if err := svc.CompleteConnect(context.Background(), "code", "never-issued-state"); err == nil {
		t.Fatal("expected an error for an unknown state")
	}
}

func TestCompleteConnectRejectsExpiredState(t *testing.T) {
	svc := NewService(testOAuthConfig("https://example.invalid/token"), newFakeRepo())
	authURL, err := svc.BeginConnect("device-1")
	if err != nil {
		t.Fatalf("BeginConnect: %v", err)
	}
	state := stateFromAuthURL(t, authURL)

	// Force it to have already expired.
	svc.mu.Lock()
	p := svc.states[state]
	p.expires = time.Now().Add(-time.Minute)
	svc.states[state] = p
	svc.mu.Unlock()

	if err := svc.CompleteConnect(context.Background(), "code", state); err == nil {
		t.Fatal("expected an error for an expired state")
	}
}

func TestAppendRowFailsWhenNotConnected(t *testing.T) {
	svc := NewService(testOAuthConfig("https://example.invalid/token"), newFakeRepo())
	err := svc.AppendRow(context.Background(), "unknown-device", "sheet-id", []string{"a", "b"})
	if err == nil {
		t.Fatal("expected an error when the device has no connection")
	}
}

func TestAppendRowSucceedsWithoutRefresh(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	sheetsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer sheetsServer.Close()

	repo := newFakeRepo()
	_ = repo.Save(domain.Connection{
		DeviceID:    "device-1",
		AccessToken: "still-valid-token",
		TokenExpiry: time.Now().Add(time.Hour), // far from expiry: no refresh call needed
	})

	svc := NewService(testOAuthConfig("https://example.invalid/token"), repo)
	svc.sheetsBaseURL = sheetsServer.URL

	err := svc.AppendRow(context.Background(), "device-1", "https://docs.google.com/spreadsheets/d/ABC123/edit#gid=0", []string{"John Doe", "john@example.com"})
	if err != nil {
		t.Fatalf("AppendRow: %v", err)
	}

	if gotAuth != "Bearer still-valid-token" {
		t.Errorf("expected the stored token to be used, got Authorization: %q", gotAuth)
	}
	if !strings.Contains(gotPath, "/v4/spreadsheets/ABC123/values/A1:append") {
		t.Errorf("expected the sheet id to be extracted from the URL, got path %q", gotPath)
	}
	values, _ := gotBody["values"].([]any)
	if len(values) != 1 {
		t.Fatalf("expected one row in the request body, got %+v", gotBody)
	}
}

func TestAppendRowRefreshesExpiredToken(t *testing.T) {
	tokenServer := fakeTokenServer(t, "refreshed-access-token")
	defer tokenServer.Close()

	var gotAuth string
	sheetsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer sheetsServer.Close()

	repo := newFakeRepo()
	_ = repo.Save(domain.Connection{
		DeviceID:     "device-1",
		AccessToken:  "expired-token",
		RefreshToken: "a-refresh-token",
		TokenExpiry:  time.Now().Add(-time.Hour), // already expired: must refresh
	})

	svc := NewService(testOAuthConfig(tokenServer.URL), repo)
	svc.sheetsBaseURL = sheetsServer.URL

	if err := svc.AppendRow(context.Background(), "device-1", "sheet-id", []string{"x"}); err != nil {
		t.Fatalf("AppendRow: %v", err)
	}

	if gotAuth != "Bearer refreshed-access-token" {
		t.Errorf("expected the refreshed token to be used, got Authorization: %q", gotAuth)
	}
	conn, _, _ := repo.FindByDevice("device-1")
	if conn.AccessToken != "refreshed-access-token" {
		t.Errorf("expected the refreshed token to be persisted, got %q", conn.AccessToken)
	}
}

func TestAppendRowSurfacesSheetsAPIErrors(t *testing.T) {
	sheetsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer sheetsServer.Close()

	repo := newFakeRepo()
	_ = repo.Save(domain.Connection{DeviceID: "device-1", AccessToken: "tok", TokenExpiry: time.Now().Add(time.Hour)})

	svc := NewService(testOAuthConfig("https://example.invalid/token"), repo)
	svc.sheetsBaseURL = sheetsServer.URL

	if err := svc.AppendRow(context.Background(), "device-1", "sheet-id", []string{"x"}); err == nil {
		t.Fatal("expected an error when the Sheets API rejects the request")
	}
}

func TestExtractSpreadsheetID(t *testing.T) {
	cases := map[string]string{
		"ABC123": "ABC123",
		"https://docs.google.com/spreadsheets/d/ABC123/edit#gid=0": "ABC123",
		"https://docs.google.com/spreadsheets/d/ABC-1_23/edit":     "ABC-1_23",
	}
	for input, want := range cases {
		if got := extractSpreadsheetID(input); got != want {
			t.Errorf("extractSpreadsheetID(%q) = %q, want %q", input, got, want)
		}
	}
}

// stateFromAuthURL pulls the `state` query param out of a generated auth URL.
func stateFromAuthURL(t *testing.T, authURL string) string {
	t.Helper()
	idx := strings.Index(authURL, "state=")
	if idx == -1 {
		t.Fatalf("no state param in %q", authURL)
	}
	rest := authURL[idx+len("state="):]
	if amp := strings.IndexByte(rest, '&'); amp != -1 {
		rest = rest[:amp]
	}
	return rest
}
