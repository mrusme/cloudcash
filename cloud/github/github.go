package github

import (
	"context"
	"time"

	"github.com/google/go-github/v85/github"
	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

type GitHub struct {
	cfg *lib.Config
	c   *github.Client
}

func New(config *lib.Config) (*GitHub, error) {
	apiKey, err := lib.Secret(
		context.Background(),
		config.Service.GitHub.APIKey,
		config.Service.GitHub.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, lib.ErrNotConfigured
	}

	s := new(GitHub)

	s.cfg = config
	s.c = github.NewClient(lib.NewHTTPClient()).WithAuthToken(apiKey)

	return s, nil
}

func (s *GitHub) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	now := time.Now().UTC()
	opts := &github.UsageReportOptions{
		Year:  github.Ptr(now.Year()),
		Month: github.Ptr(int(now.Month())),
	}

	charges := decimal.Zero

	for _, user := range s.cfg.Service.GitHub.Users {
		report, _, err := s.c.Billing.GetUsageReport(ctx, user, opts)
		if err != nil {
			return nil, err
		}
		charges = charges.Add(netAmount(report))
	}

	for _, org := range s.cfg.Service.GitHub.Orgs {
		report, _, err := s.c.Billing.GetOrganizationUsageReport(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		charges = charges.Add(netAmount(report))
	}

	status := new(lib.ServiceStatus)

	status.Currency = "USD"
	status.CurrentCharges = charges.RoundBank(2)

	return status, nil
}

func netAmount(report *github.UsageReport) decimal.Decimal {
	total := decimal.Zero

	for _, item := range report.UsageItems {
		total = total.Add(decimal.NewFromFloat(item.NetAmount))
	}

	return total
}
