package digitalocean

import (
	"context"

	"github.com/digitalocean/godo"
	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type DigitalOcean struct {
	c *godo.Client
}

func New(config *lib.Config) (*DigitalOcean, error) {
	apiKey, err := lib.Secret(
		context.Background(),
		config.Service.DigitalOcean.APIKey,
		config.Service.DigitalOcean.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, lib.ErrNotConfigured
	}

	s := new(DigitalOcean)
	s.c = godo.NewClient(lib.NewBearerHTTPClient(apiKey))

	return s, nil
}

func (s *DigitalOcean) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	balance, _, err := s.c.Balance.Get(ctx)
	if err != nil {
		return nil, err
	}

	status := new(lib.ServiceStatus)

	status.Currency = "USD"

	status.AccountBalance, err = decimal.NewFromString(balance.AccountBalance)
	if err != nil {
		return nil, err
	}

	status.CurrentCharges, err = decimal.NewFromString(balance.MonthToDateUsage)
	if err != nil {
		return nil, err
	}

	return status, nil
}
