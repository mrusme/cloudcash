package vultr

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/vultr/govultr/v3"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type Vultr struct {
	cfg *lib.Config
	c   *govultr.Client
}

func New(config *lib.Config) (*Vultr, error) {
	if !lib.HasSecret(
		config.Service.Vultr.APIKey,
		config.Service.Vultr.APIKeyCommand,
	) {
		return nil, lib.ErrNotConfigured
	}

	s := new(Vultr)

	s.cfg = config

	return s, nil
}

func (s *Vultr) client(ctx context.Context) (*govultr.Client, error) {
	if s.c != nil {
		return s.c, nil
	}

	apiKey, err := lib.Secret(
		ctx,
		s.cfg.Service.Vultr.APIKey,
		s.cfg.Service.Vultr.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}

	s.c = govultr.NewClient(lib.NewBearerHTTPClient(apiKey))
	s.c.SetUserAgent(lib.UserAgent)

	return s.c, nil
}

func (s *Vultr) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	c, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	account, _, err := c.Account.Get(ctx)
	if err != nil {
		return nil, err
	}

	status := new(lib.ServiceStatus)

	status.Currency = "USD"
	status.AccountBalance = decimal.NewFromFloat32(-account.Balance)
	status.CurrentCharges = decimal.NewFromFloat32(account.PendingCharges)
	status.PreviousCharges = decimal.NewFromFloat32(-account.LastPaymentAmount)

	return status, nil
}
