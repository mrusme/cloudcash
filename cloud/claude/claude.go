package claude

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "net/http"
  "os"
  "path/filepath"
  "time"

  "github.com/shopspring/decimal"

  "xn--gckvb8fzb.com/cloudcash/lib"
)

// Anthropic does not offer a documented API for subscription (Pro/Max) usage.
// This is the same endpoint the Claude Code CLI uses for its `/usage` command;
// it is undocumented and may change without notice.
const ENDPOINT string = "https://api.anthropic.com/api/oauth/usage"
const OAUTH_BETA string = "oauth-2025-04-20"
const USER_AGENT string = "xn--gckvb8fzb.com/cloudcash"

// The endpoint aggressively rate limits requests without a User-Agent.
const TIMEOUT time.Duration = 15 * time.Second

type Claude struct {
  cfg        *lib.Config
  c          *http.Client
}

type credentials struct {
  ClaudeAiOauth      struct {
    AccessToken      string `json:"accessToken"`
    ExpiresAt        int64  `json:"expiresAt"`
  } `json:"claudeAiOauth"`
}

type window struct {
  Utilization      float64 `json:"utilization"`
}

type money struct {
  AmountMinor      int64   `json:"amount_minor"`
  Exponent         int32   `json:"exponent"`
}

type usage struct {
  FiveHour           window  `json:"five_hour"`
  SevenDay           window  `json:"seven_day"`
  Spend              struct {
    Used            *money   `json:"used"`
  } `json:"spend"`
  ExtraUsage         struct {
    UsedCredits     *float64 `json:"used_credits"`
    DecimalPlaces   *int32   `json:"decimal_places"`
  } `json:"extra_usage"`
}

func New(config *lib.Config) (*Claude, error) {
  if config.Service.Claude.Enabled == false {
    return nil, errors.New("Not enabled")
  }

  s := new(Claude)

  s.cfg = config
  s.c = &http.Client{Timeout: TIMEOUT}

  return s, nil
}

func (s *Claude) GetServiceStatus() (*lib.ServiceStatus, error) {
  token, err := s.token()
  if err != nil {
    return nil, err
  }

  u, err := s.fetch(token)
  if err != nil {
    return nil, err
  }

  status := new(lib.ServiceStatus)

  status.AccountBalance = decimal.NewFromInt(0)
  status.CurrentCharges = spent(u)
  status.PreviousCharges = decimal.NewFromInt(0)
  status.SessionUsage = decimal.NewFromFloat(u.FiveHour.Utilization)
  status.WeeklyUsage = decimal.NewFromFloat(u.SevenDay.Utilization)

  return status, nil
}

// token returns the OAuth access token, either straight from the configuration
// or from the credentials file the Claude Code CLI maintains.
func (s *Claude) token() (string, error) {
  if s.cfg.Service.Claude.OAuthToken != "" {
    return s.cfg.Service.Claude.OAuthToken, nil
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
    return "", fmt.Errorf("No access token in %s", path)
  }

  // ExpiresAt is a Unix timestamp in milliseconds. Refreshing the token is up
  // to the Claude Code CLI, cloudcash only reads what's on disk.
  if creds.ClaudeAiOauth.ExpiresAt > 0 &&
     time.UnixMilli(creds.ClaudeAiOauth.ExpiresAt).Before(time.Now()) {
    return "", errors.New("Access token expired")
  }

  return creds.ClaudeAiOauth.AccessToken, nil
}

func (s *Claude) fetch(token string) (*usage, error) {
  ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
  defer cancel()

  req, err := http.NewRequestWithContext(ctx, http.MethodGet, ENDPOINT, nil)
  if err != nil {
    return nil, err
  }

  req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
  req.Header.Set("anthropic-beta", OAUTH_BETA)
  req.Header.Set("User-Agent", USER_AGENT)

  resp, err := s.c.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("%s returned %s", ENDPOINT, resp.Status)
  }

  u := new(usage)
  if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
    return nil, err
  }

  return u, nil
}

// spent returns the usage credits consumed so far. `spend.used` is the current
// shape; `extra_usage.used_credits` is the older one and is kept as a fallback.
func spent(u *usage) (decimal.Decimal) {
  if u.Spend.Used != nil {
    return decimal.New(u.Spend.Used.AmountMinor, -u.Spend.Used.Exponent)
  }

  if u.ExtraUsage.UsedCredits != nil {
    var places int32 = 0
    if u.ExtraUsage.DecimalPlaces != nil {
      places = *u.ExtraUsage.DecimalPlaces
    }
    return decimal.NewFromFloat(
      *u.ExtraUsage.UsedCredits / math.Pow(10, float64(places)),
    )
  }

  return decimal.NewFromInt(0)
}
