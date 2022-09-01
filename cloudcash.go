package main

import (
  "context"
  "encoding/json"
  "fmt"

  "github.com/vultr/govultr/v2"
  "golang.org/x/oauth2"
)

type Services struct {
  Vultr struct {
    ctx       context.Context
    oauth2cfg oauth2.Config
    c         *govultr.Client
    Status    ServiceStatus `json:"status,omitempty"`
  } `json:"vultr,omitempty"`
}

type ServiceStatus struct {
  CurrentCost float32 `json:"current_cost"`
}

func main() {
  config, err := Cfg()
  if err != nil {
    panic(err)
  }

  services := new(Services)
  services.initVultr(&config)
  services.refreshVultr()

  outputJson, _ := json.Marshal(services)
  fmt.Println(string(outputJson))

  return
}

func (s *Services) initVultr(config *Config) (error) {
  s.Vultr.ctx = context.Background()
  s.Vultr.oauth2cfg = oauth2.Config{}
  ts := s.Vultr.oauth2cfg.TokenSource(s.Vultr.ctx, &oauth2.Token{AccessToken: config.Service.Vultr.APIKey})
  s.Vultr.c = govultr.NewClient(oauth2.NewClient(s.Vultr.ctx, ts))
  s.Vultr.c.SetUserAgent("github.com/mrusme/cloudcash")
  return nil
}

func (s *Services) refreshVultr() (error) {
  account, err := s.Vultr.c.Account.Get(s.Vultr.ctx)
  if err != nil {
    return err
  }

  s.Vultr.Status.CurrentCost = account.PendingCharges
  return nil
}
