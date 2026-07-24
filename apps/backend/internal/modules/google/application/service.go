// Package application orchestrates Charli's Google integration (L4): the
// OAuth connect/callback flow, and performing a confirmed sheets_append
// action against the Sheets API using the stored, auto-refreshed token.
package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/levelaxis/charli/backend/internal/modules/google/domain"
)

// stateTTL bounds how long an OAuth "state" (linking a connect attempt back
// to the device that started it) stays valid.
const stateTTL = 10 * time.Minute

const defaultSheetsBaseURL = "https://sheets.googleapis.com"

var spreadsheetURLPattern = regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9_-]+)`)

// pendingState is one in-flight OAuth attempt, keyed by the random state
// value sent to Google and back.
type pendingState struct {
	deviceID string
	expires  time.Time
}

// Service is nil-safe by design: without both a repository (no database
// configured) and an oauth config (no GOOGLE_CLIENT_ID/SECRET set), the
// integration is simply unavailable — every method reports that clearly
// rather than behaving inconsistently.
type Service struct {
	oauth         *oauth2.Config
	repo          domain.Repository
	sheetsBaseURL string

	mu     sync.Mutex
	states map[string]pendingState
}

// NewService builds a Service. oauthConfig and repo may both be nil, in
// which case the integration reports itself unavailable.
func NewService(oauthConfig *oauth2.Config, repo domain.Repository) *Service {
	return &Service{
		oauth:         oauthConfig,
		repo:          repo,
		sheetsBaseURL: defaultSheetsBaseURL,
		states:        make(map[string]pendingState),
	}
}

// Available reports whether the integration can be used at all.
func (s *Service) Available() bool {
	return s.oauth != nil && s.repo != nil
}

// BeginConnect starts an OAuth attempt for deviceID and returns the Google
// consent URL to open.
func (s *Service) BeginConnect(deviceID string) (string, error) {
	if !s.Available() {
		return "", errors.New("google integration is not configured")
	}

	state, err := randomState()
	if err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}

	s.mu.Lock()
	s.states[state] = pendingState{deviceID: deviceID, expires: time.Now().Add(stateTTL)}
	s.mu.Unlock()

	authURL := s.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	return authURL, nil
}

// CompleteConnect finishes an OAuth attempt: validates state, exchanges code
// for tokens, and persists the connection.
func (s *Service) CompleteConnect(ctx context.Context, code, state string) error {
	if !s.Available() {
		return errors.New("google integration is not configured")
	}

	s.mu.Lock()
	pending, ok := s.states[state]
	if ok {
		delete(s.states, state)
	}
	s.mu.Unlock()

	if !ok || time.Now().After(pending.expires) {
		return errors.New("invalid or expired state")
	}

	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("exchange code: %w", err)
	}

	return s.repo.Save(domain.Connection{
		DeviceID:     pending.deviceID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
	})
}

// IsConnected reports whether deviceID has a completed Google connection.
func (s *Service) IsConnected(deviceID string) bool {
	if !s.Available() {
		return false
	}
	_, found, err := s.repo.FindByDevice(deviceID)
	return err == nil && found
}

// AppendRow appends values as a new row in spreadsheetIDOrURL (a raw sheet
// ID, or a full Google Sheets URL — either is accepted), using deviceID's
// stored, auto-refreshed token.
func (s *Service) AppendRow(ctx context.Context, deviceID, spreadsheetIDOrURL string, values []string) error {
	if !s.Available() {
		return errors.New("google sheets isn't connected yet")
	}

	conn, found, err := s.repo.FindByDevice(deviceID)
	if err != nil {
		return fmt.Errorf("look up connection: %w", err)
	}
	if !found {
		return errors.New("google sheets isn't connected yet")
	}

	token := &oauth2.Token{
		AccessToken:  conn.AccessToken,
		RefreshToken: conn.RefreshToken,
		Expiry:       conn.TokenExpiry,
	}

	fresh, err := s.oauth.TokenSource(ctx, token).Token()
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}
	if fresh.AccessToken != token.AccessToken {
		// Best-effort: if persisting the refreshed token fails, we still
		// proceed with it for this call — worst case we refresh again next time.
		_ = s.repo.Save(domain.Connection{
			DeviceID:     deviceID,
			AccessToken:  fresh.AccessToken,
			RefreshToken: fresh.RefreshToken,
			TokenExpiry:  fresh.Expiry,
		})
	}

	sheetID := extractSpreadsheetID(spreadsheetIDOrURL)
	body, err := json.Marshal(map[string]any{"values": [][]string{values}})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v4/spreadsheets/%s/values/A1:append?valueInputOption=USER_ENTERED", s.sheetsBaseURL, sheetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.oauth.Client(ctx, fresh).Do(req)
	if err != nil {
		return fmt.Errorf("call sheets api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sheets api returned status %d", resp.StatusCode)
	}
	return nil
}

// extractSpreadsheetID returns s unchanged if it looks like a raw ID, or the
// ID portion if s is a full Google Sheets URL.
func extractSpreadsheetID(s string) string {
	if m := spreadsheetURLPattern.FindStringSubmatch(s); len(m) == 2 {
		return m[1]
	}
	return s
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
