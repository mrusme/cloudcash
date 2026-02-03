package vultr

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/vultr/govultr/v3"

	"golang.org/x/oauth2"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type Vultr struct {
	ctx       context.Context
	oauth2cfg oauth2.Config
	c         *govultr.Client
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
	s.c.SetUserAgent("xn--gckvb8fzb.com/cloudcash")

	return s, nil
}

func (s *Vultr) GetServiceStatus() (*lib.ServiceStatus, error) {
	account, _, err := s.c.Account.Get(s.ctx)
	if err != nil {
		return nil, err
	}

	status := new(lib.ServiceStatus)

	status.AccountBalance = decimal.NewFromFloat32(account.Balance * -1.0)
	status.CurrentCharges = decimal.NewFromFloat32(account.PendingCharges)
	status.PreviousCharges = decimal.NewFromFloat32(account.LastPaymentAmount).
		Mul(decimal.NewFromInt(-1))

	return status, nil
}
