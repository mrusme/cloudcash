package vultr

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/vultr/govultr/v3"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type Vultr struct {
	c *govultr.Client
}

func New(config *lib.Config) (*Vultr, error) {
	apiKey, err := lib.Secret(
		context.Background(),
		config.Service.Vultr.APIKey,
		config.Service.Vultr.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, lib.ErrNotConfigured
	}

	s := new(Vultr)

	s.c = govultr.NewClient(lib.NewBearerHTTPClient(apiKey))
	s.c.SetUserAgent(lib.UserAgent)

	return s, nil
}

func (s *Vultr) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	account, _, err := s.c.Account.Get(ctx)
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
