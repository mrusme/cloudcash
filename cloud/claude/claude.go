package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

// Anthropic does not offer a documented API for subscription (Pro/Max) usage.
// This is the same endpoint the Claude Code CLI uses for its `/usage` command;
// it is undocumented and may change without notice.
const endpoint = "https://api.anthropic.com/api/oauth/usage"
const oauthBeta = "oauth-2025-04-20"

type Claude struct {
	cfg *lib.Config
	c   *http.Client
}

type credentials struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"`
	} `json:"claudeAiOauth"`
}

type window struct {
	Utilization float64   `json:"utilization"`
	ResetsAt    time.Time `json:"resets_at"`
}

type usage struct {
	FiveHour   window `json:"five_hour"`
	SevenDay   window `json:"seven_day"`
	ExtraUsage struct {
		UsedCredits *float64 `json:"used_credits"`
		Currency    string   `json:"currency"`
	} `json:"extra_usage"`
}

func New(config *lib.Config) (*Claude, error) {
	if !config.Service.Claude.Enabled {
		return nil, lib.ErrNotConfigured
	}

	s := new(Claude)

	s.cfg = config
	s.c = lib.NewHTTPClient()

	return s, nil
}

func (s *Claude) UsageOnly() bool {
	return s.cfg.Service.Claude.UsageOnly
}

func (s *Claude) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	token, err := s.token(ctx)
	if err != nil {
		return nil, err
	}

	u := new(usage)
	err = lib.GetJSON(ctx, s.c, endpoint, map[string]string{
		"Authorization":  "Bearer " + token,
		"anthropic-beta": oauthBeta,
	}, u)
	if err != nil {
		return nil, err
	}

	status := new(lib.ServiceStatus)

	status.CurrentCharges, status.Currency = spent(u)
	status.SessionUsage = decimal.NewFromFloat(u.FiveHour.Utilization)
	status.SessionResetsAt = u.FiveHour.ResetsAt.UTC()
	status.WeeklyUsage = decimal.NewFromFloat(u.SevenDay.Utilization)
	status.WeeklyResetsAt = u.SevenDay.ResetsAt.UTC()

	return status, nil
}

func (s *Claude) token(ctx context.Context) (string, error) {
	token, err := lib.Secret(
		ctx,
		s.cfg.Service.Claude.OAuthToken,
		s.cfg.Service.Claude.OAuthTokenCommand,
	)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}

	path := s.cfg.Service.Claude.CredentialsFile
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, ".claude", ".credentials.json")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var creds credentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return "", err
	}

	if creds.ClaudeAiOauth.AccessToken == "" {
		return "", fmt.Errorf("no access token in %s", path)
	}

	// ExpiresAt is a Unix timestamp in milliseconds. Refreshing the token is up
	// to the Claude Code CLI, cloudcash only reads what's on disk.
	if creds.ClaudeAiOauth.ExpiresAt > 0 &&
		time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt).Before(time.Now()) {
		return "", errors.New("access token expired")
	}

	return creds.ClaudeAiOauth.AccessToken, nil
}

func spent(u *usage) (decimal.Decimal, string) {
	if u.ExtraUsage.UsedCredits == nil {
		return decimal.Zero, ""
	}

	currency := strings.ToUpper(u.ExtraUsage.Currency)
	if currency == "" {
		currency = "USD"
	}

	var places int32 = 2
	switch currency {
	case "JPY", "KRW", "VND":
		places = 0
	}

	return decimal.NewFromFloat(*u.ExtraUsage.UsedCredits).Shift(-places), currency
}
