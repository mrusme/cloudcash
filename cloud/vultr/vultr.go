package vultr

import (
  "context"
  "errors"

  "github.com/shopspring/decimal"
  "github.com/vultr/govultr/v2"

  "golang.org/x/oauth2"

  "github.com/mrusme/cloudcash/lib"
)

type Vultr struct {
  ctx        context.Context
  oauth2cfg  oauth2.Config
  c          *govultr.Client
}

func New(config *lib.Config) (*Vultr, error) {
  if config.Service.Vultr.APIKey == "" {
    return nil, errors.New("No API key")
  }

  s := new(Vultr)

  s.ctx = context.Background()
  s.oauth2cfg = oauth2.Config{}
  ts := s.oauth2cfg.TokenSource(s.ctx, &oauth2.Token{AccessToken: config.Service.Vultr.APIKey})
  s.c = govultr.NewClient(oauth2.NewClient(s.ctx, ts))
  s.c.SetUserAgent("github.com/mrusme/cloudcash")

  return s, nil
}

func (s *Vultr) GetServiceStatus() (*lib.ServiceStatus, error) {
  account, err := s.c.Account.Get(s.ctx)
  if err != nil {
    return nil, err
  }

  status := new(lib.ServiceStatus)

  status.AccountBalance = decimal.NewFromFloat32(account.Balance)
  status.CurrentCharges = decimal.NewFromFloat32(account.PendingCharges)
  status.PreviousCharges = decimal.NewFromFloat32(account.LastPaymentAmount).
                            Mul(decimal.NewFromInt(-1))

  return status, nil
}

