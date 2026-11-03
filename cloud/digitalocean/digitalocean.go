package digitalocean

import (
	"context"

	"github.com/digitalocean/godo"
	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type DigitalOcean struct {
	cfg *lib.Config
	c   *godo.Client
}

func New(config *lib.Config) (*DigitalOcean, error) {
	if !lib.HasSecret(
		config.Service.DigitalOcean.APIKey,
		config.Service.DigitalOcean.APIKeyCommand,
	) {
		return nil, lib.ErrNotConfigured
	}

	s := new(DigitalOcean)
	s.cfg = config

	return s, nil
}

func (s *DigitalOcean) client(ctx context.Context) (*godo.Client, error) {
	if s.c != nil {
		return s.c, nil
	}

	apiKey, err := lib.Secret(
		ctx,
		s.cfg.Service.DigitalOcean.APIKey,
		s.cfg.Service.DigitalOcean.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}

	s.c = godo.NewClient(lib.NewBearerHTTPClient(apiKey))

	return s.c, nil
}

func (s *DigitalOcean) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	c, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	balance, _, err := c.Balance.Get(ctx)
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
