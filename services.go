package main

import (
  "context"
  "fmt"
  "bytes"
  "errors"

  "github.com/vultr/govultr/v2"
  "github.com/digitalocean/godo"

  "github.com/shopspring/decimal"
  "golang.org/x/oauth2"
  "text/template"
)

type Services struct {
  Vultr        struct {
    ready      bool
    ctx        context.Context
    oauth2cfg  oauth2.Config
    c          *govultr.Client
    Status     ServiceStatus `json:"status,omitempty"`
  } `json:"vultr,omitempty"`
  DigitalOcean struct {
    ready      bool
    c          *godo.Client
    Status     ServiceStatus `json:"status,omitempty"`
  } `json:"digitalocean,omitempty"`
}

type ServiceStatus struct {
  Service         string          `json:"service"`
  AccountBalance  decimal.Decimal `json:"account_balance"`
  CurrentCharges  decimal.Decimal `json:"current_charges"`
  PreviousCharges decimal.Decimal `json:"previous_charges"`
}

type WaybarOutput struct {
  Text    string `json:"text"`
  Tooltip string `json:"tooltip"`
  Alt     string `json:"alt"`
  Class   string `json:"class"`
}

func (s *Services) initAll(config *Config) () {
  if s.initVultr(config) != nil {
    s.Vultr.ready = false
  } else {
    s.Vultr.ready = true
  }

  if s.initDigitalOcean(config) != nil {
    s.DigitalOcean.ready = false
  } else {
    s.DigitalOcean.ready = true
  }

  return
}

func (s *Services) refreshAll() () {
  s.refreshVultr()
  s.refreshDigitalOcean()
  return
}

func (s *Services) initVultr(config *Config) (error) {
  if config.Service.Vultr.APIKey == "" {
    return errors.New("No API key")
  }

  s.Vultr.ctx = context.Background()
  s.Vultr.oauth2cfg = oauth2.Config{}
  ts := s.Vultr.oauth2cfg.TokenSource(s.Vultr.ctx, &oauth2.Token{AccessToken: config.Service.Vultr.APIKey})
  s.Vultr.c = govultr.NewClient(oauth2.NewClient(s.Vultr.ctx, ts))
  s.Vultr.c.SetUserAgent("github.com/mrusme/cloudcash")
  return nil
}

func (s *Services) refreshVultr() (error) {
  if s.Vultr.ready == false {
    return errors.New("Not ready")
  }

  account, err := s.Vultr.c.Account.Get(s.Vultr.ctx)
  if err != nil {
    return err
  }

  s.Vultr.Status.Service = "Vultr"
  s.Vultr.Status.AccountBalance = decimal.NewFromFloat32(account.Balance)
  s.Vultr.Status.CurrentCharges = decimal.NewFromFloat32(account.PendingCharges)
  s.Vultr.Status.PreviousCharges = decimal.NewFromFloat32(account.LastPaymentAmount)
  return nil
}

func (s *Services) initDigitalOcean(config *Config) (error) {
  if config.Service.DigitalOcean.APIKey == "" {
    return errors.New("No API key")
  }

  s.DigitalOcean.c = godo.NewFromToken(config.Service.DigitalOcean.APIKey)
  return nil
}

func (s *Services) refreshDigitalOcean() (error) {
  if s.DigitalOcean.ready == false {
    return errors.New("Not ready")
  }

  ctx := context.Background()
  balance, _, err := s.DigitalOcean.c.Balance.Get(ctx)
  if err != nil {
    return err
  }

  s.DigitalOcean.Status.Service = "DigitalOcean"
  s.DigitalOcean.Status.AccountBalance, _ = decimal.NewFromString(balance.AccountBalance)
  s.DigitalOcean.Status.CurrentCharges, _ = decimal.NewFromString(balance.MonthToDateUsage)
  s.DigitalOcean.Status.PreviousCharges, _ = decimal.NewFromString("0.0")
  return nil
}

func (s *Services) toJSON() (string) {
  outputJson, _ := JSONMarshal(s)
  return string(outputJson)
}

func (s *Services) toWaybar(t *template.Template) (string) {
  waybarOutput := new(WaybarOutput)

  var vultr bytes.Buffer
  if err := t.Execute(&vultr, s.Vultr.Status); err != nil {
    panic(err)
  }

  var digitalocean bytes.Buffer
  if err := t.Execute(&digitalocean, s.DigitalOcean.Status); err != nil {
    panic(err)
  }

  waybarOutput.Text = fmt.Sprintf(
    "%s %s",
    vultr.String(),
    digitalocean.String(),
  )

  outputJson, _ := JSONMarshal(waybarOutput)
  return string(outputJson)
}

