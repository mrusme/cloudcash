package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

// OpenAI does not offer a documented API for ChatGPT/Codex subscription usage.
// This is the endpoint the Codex CLI polls for its `/status` output; it is
// undocumented and may change without notice.
const endpoint = "https://chatgpt.com/backend-api/wham/usage"

type Codex struct {
	cfg *lib.Config
	c   *http.Client
}

// $CODEX_HOME/auth.json, as written by the Codex CLI.
type credentials struct {
	Tokens *struct {
		AccessToken string `json:"access_token"`
		AccountID   string `json:"account_id"`
	} `json:"tokens"`
}

type window struct {
	UsedPercent float64 `json:"used_percent"`
	ResetAt     int64   `json:"reset_at"`
}

func (w *window) resetsAt() time.Time {
	if w.ResetAt <= 0 {
		return time.Time{}
	}

	return time.Unix(w.ResetAt, 0).UTC()
}

type usage struct {
	RateLimit *struct {
		PrimaryWindow   *window `json:"primary_window"`
		SecondaryWindow *window `json:"secondary_window"`
	} `json:"rate_limit"`
	Credits *struct {
		Unlimited bool    `json:"unlimited"`
		Balance   *string `json:"balance"`
	} `json:"credits"`
}

func New(config *lib.Config) (*Codex, error) {
	if !config.Service.Codex.Enabled {
		return nil, lib.ErrNotConfigured
	}

	s := new(Codex)

	s.cfg = config
	s.c = lib.NewHTTPClient()

	return s, nil
}

func (s *Codex) UsageOnly() bool {
	return s.cfg.Service.Codex.UsageOnly
}

func (s *Codex) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	token, account, err := s.credentials(ctx)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{"Authorization": "Bearer " + token}
	if account != "" {
		headers["ChatGPT-Account-Id"] = account
	}

	u := new(usage)
	if err := lib.GetJSON(ctx, s.c, endpoint, headers, u); err != nil {
		return nil, err
	}

	status := new(lib.ServiceStatus)

	// The endpoint reports credits remaining, not credits spent, so there is
	// nothing to put into CurrentCharges.
	status.AccountBalance = balance(u)

	if u.RateLimit != nil {
		if u.RateLimit.PrimaryWindow != nil {
			status.SessionUsage = decimal.NewFromFloat(
				u.RateLimit.PrimaryWindow.UsedPercent,
			)
			status.SessionResetsAt = u.RateLimit.PrimaryWindow.resetsAt()
		}
		if u.RateLimit.SecondaryWindow != nil {
			status.WeeklyUsage = decimal.NewFromFloat(
				u.RateLimit.SecondaryWindow.UsedPercent,
			)
			status.WeeklyResetsAt = u.RateLimit.SecondaryWindow.resetsAt()
		}
	}

	return status, nil
}

func (s *Codex) credentials(ctx context.Context) (string, string, error) {
	token, err := lib.Secret(
		ctx,
		s.cfg.Service.Codex.OAuthToken,
		s.cfg.Service.Codex.OAuthTokenCommand,
	)
	if err != nil {
		return "", "", err
	}
	if token != "" {
		return token, s.cfg.Service.Codex.AccountID, nil
	}

	path, err := s.credentialsFile()
	if err != nil {
		return "", "", err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	var creds credentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return "", "", err
	}

	if creds.Tokens == nil || creds.Tokens.AccessToken == "" {
		return "", "", fmt.Errorf("no access token in %s", path)
	}

	account := s.cfg.Service.Codex.AccountID
	if account == "" {
		account = creds.Tokens.AccountID
	}

	return creds.Tokens.AccessToken, account, nil
}

func (s *Codex) credentialsFile() (string, error) {
	if s.cfg.Service.Codex.CredentialsFile != "" {
		return s.cfg.Service.Codex.CredentialsFile, nil
	}

	home := os.Getenv("CODEX_HOME")
	if home == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(dir, ".codex")
	}

	return filepath.Join(home, "auth.json"), nil
}

// balance returns the credits left to spend. Accounts on an unlimited plan
// report no meaningful figure.
func balance(u *usage) decimal.Decimal {
	if u.Credits == nil ||
		u.Credits.Unlimited ||
		u.Credits.Balance == nil {
		return decimal.Zero
	}

	b, err := decimal.NewFromString(*u.Credits.Balance)
	if err != nil {
		return decimal.Zero
	}

	return b
}
