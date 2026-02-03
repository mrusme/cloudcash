package digitalocean

import (
  "context"
  "errors"

  "github.com/shopspring/decimal"
  "github.com/digitalocean/godo"

  "xn--gckvb8fzb.com/cloudcash/lib"
)

type DigitalOcean struct {
  c          *godo.Client
}

func New(config *lib.Config) (*DigitalOcean, error) {
  if config.Service.DigitalOcean.APIKey == "" {
    return nil, errors.New("No API key")
  }

  s := new(DigitalOcean)
  s.c = godo.NewFromToken(config.Service.DigitalOcean.APIKey)

  return s, nil
}

func (s *DigitalOcean) GetServiceStatus() (*lib.ServiceStatus, error) {
  ctx := context.Background()
  balance, _, err := s.c.Balance.Get(ctx)
  if err != nil {
    return nil, err
  }

  status := new(lib.ServiceStatus)

  status.AccountBalance, _ = decimal.NewFromString(balance.AccountBalance)
  status.CurrentCharges, _ = decimal.NewFromString(balance.MonthToDateUsage)
  status.PreviousCharges, _ = decimal.NewFromString("0.0")

  return status, nil
}


