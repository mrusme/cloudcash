package github

import (
  "context"
  "errors"

  "github.com/shopspring/decimal"
  "github.com/google/go-github/v47/github"

  "golang.org/x/oauth2"

  "github.com/mrusme/cloudcash/lib"
)

var PRICE_PER_MINUTE decimal.Decimal = decimal.NewFromFloat32(0.008)
var PRICE_PER_GB decimal.Decimal = decimal.NewFromFloat32(0.008)

type GitHub struct {
  cfg        *lib.Config
  ctx        context.Context
  oauth2cfg  oauth2.Config
  c          *github.Client
}

func New(config *lib.Config) (*GitHub, error) {
  if config.Service.GitHub.APIKey == "" {
    return nil, errors.New("No API key")
  }

  s := new(GitHub)

  s.cfg = config
  s.ctx = context.Background()
  s.oauth2cfg = oauth2.Config{}
  ts := s.oauth2cfg.TokenSource(s.ctx, &oauth2.Token{AccessToken: config.Service.GitHub.APIKey})
  s.c = github.NewClient(oauth2.NewClient(s.ctx, ts))

  return s, nil
}

func (s *GitHub) GetServiceStatus() (*lib.ServiceStatus, error) {
  ctx := context.Background()

  var currentCharges decimal.Decimal = decimal.NewFromInt(0)

  for _, user := range s.cfg.Service.GitHub.Users {
    actionBilling, _, err := s.c.Billing.GetActionsBillingUser(ctx, user)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      actionBilling.TotalPaidMinutesUsed,
      PRICE_PER_MINUTE,
    )

    packagesBilling, _, err := s.c.Billing.GetPackagesBillingUser(ctx, user)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      packagesBilling.TotalPaidGigabytesBandwidthUsed,
      PRICE_PER_GB,
    )

    storageBilling, _, err := s.c.Billing.GetStorageBillingUser(ctx, user)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      storageBilling.EstimatedPaidStorageForMonth,
      PRICE_PER_GB,
    )
  }

  for _, org := range s.cfg.Service.GitHub.Orgs {
    actionBilling, _, err := s.c.Billing.GetActionsBillingOrg(ctx, org)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      actionBilling.TotalPaidMinutesUsed,
      PRICE_PER_MINUTE,
    )

    packagesBilling, _, err := s.c.Billing.GetPackagesBillingOrg(ctx, org)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      packagesBilling.TotalPaidGigabytesBandwidthUsed,
      PRICE_PER_GB,
    )

    storageBilling, _, err := s.c.Billing.GetStorageBillingOrg(ctx, org)
    if err != nil {
      // TODO: Handle error
      continue
    }
    currentCharges = AddTo(
      currentCharges,
      storageBilling.EstimatedPaidStorageForMonth,
      PRICE_PER_GB,
    )
  }

  status := new(lib.ServiceStatus)

  status.AccountBalance = decimal.NewFromInt(0)
  status.CurrentCharges = currentCharges.RoundBank(2)
  status.PreviousCharges = decimal.NewFromInt(0)

  return status, nil
}

func AddTo(dec decimal.Decimal, i interface{}, mul decimal.Decimal) (decimal.Decimal) {
  switch x := i.(type) {
    case float64:
      return dec.Add(
        decimal.NewFromFloat(x).
          Mul(mul),
      )
    case int:
      return dec.Add(
        decimal.NewFromInt(int64(x)).
          Mul(mul),
      )
    default:
      return decimal.Decimal{}
  }
}

