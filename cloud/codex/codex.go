package codex

import (
  "context"
  "encoding/json"
  "errors"
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
const ENDPOINT string = "https://chatgpt.com/backend-api/wham/usage"
const USER_AGENT string = "xn--gckvb8fzb.com/cloudcash"

const TIMEOUT time.Duration = 15 * time.Second

type Codex struct {
  cfg                  *lib.Config
  c                    *http.Client
}

// $CODEX_HOME/auth.json, as written by the Codex CLI.
type credentials struct {
  Tokens               *struct {
    AccessToken         string `json:"access_token"`
    AccountID           string `json:"account_id"`
  } `json:"tokens"`
}

type window struct {
  UsedPercent           float64 `json:"used_percent"`
}

type usage struct {
  RateLimit            *struct {
    PrimaryWindow      *window `json:"primary_window"`
    SecondaryWindow    *window `json:"secondary_window"`
  } `json:"rate_limit"`
  Credits              *struct {
    Unlimited           bool    `json:"unlimited"`
    Balance            *string  `json:"balance"`
  } `json:"credits"`
}

func New(config *lib.Config) (*Codex, error) {
  if config.Service.Codex.Enabled == false {
    return nil, errors.New("Not enabled")
  }

  s := new(Codex)

  s.cfg = config
  s.c = &http.Client{Timeout: TIMEOUT}

  return s, nil
}

func (s *Codex) GetServiceStatus() (*lib.ServiceStatus, error) {
  token, account, err := s.credentials()
  if err != nil {
    return nil, err
  }

  u, err := s.fetch(token, account)
  if err != nil {
    return nil, err
  }

  status := new(lib.ServiceStatus)

  // The endpoint reports credits remaining, not credits spent, so there is
  // nothing to put into CurrentCharges.
  status.AccountBalance = balance(u)
  status.CurrentCharges = decimal.NewFromInt(0)
  status.PreviousCharges = decimal.NewFromInt(0)

  if u.RateLimit != nil {
    if u.RateLimit.PrimaryWindow != nil {
      status.SessionUsage = decimal.NewFromFloat(
        u.RateLimit.PrimaryWindow.UsedPercent,
      )
    }
    if u.RateLimit.SecondaryWindow != nil {
      status.WeeklyUsage = decimal.NewFromFloat(
        u.RateLimit.SecondaryWindow.UsedPercent,
      )
    }
  }

  return status, nil
}

// credentials returns the OAuth access token and the ChatGPT account ID, either
// straight from the configuration or from the auth file the Codex CLI writes.
func (s *Codex) credentials() (string, string, error) {
  if s.cfg.Service.Codex.OAuthToken != "" {
    return s.cfg.Service.Codex.OAuthToken,
           s.cfg.Service.Codex.AccountID,
           nil
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
    return "", "", fmt.Errorf("No access token in %s", path)
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

func (s *Codex) fetch(token string, account string) (*usage, error) {
  ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
  defer cancel()

  req, err := http.NewRequestWithContext(ctx, http.MethodGet, ENDPOINT, nil)
  if err != nil {
    return nil, err
  }

  req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
  req.Header.Set("User-Agent", USER_AGENT)
  if account != "" {
    req.Header.Set("ChatGPT-Account-Id", account)
  }

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

// balance returns the credits left to spend. Accounts on an unlimited plan
// report no meaningful figure.
func balance(u *usage) (decimal.Decimal) {
  if u.Credits == nil ||
     u.Credits.Unlimited == true ||
     u.Credits.Balance == nil {
    return decimal.NewFromInt(0)
  }

  b, err := decimal.NewFromString(*u.Credits.Balance)
  if err != nil {
    return decimal.NewFromInt(0)
  }

  return b
}
