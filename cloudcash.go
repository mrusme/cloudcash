package main

import (
  "flag"
  "context"
  "encoding/json"
  "fmt"
  "bytes"

  "github.com/vultr/govultr/v2"
  "golang.org/x/oauth2"
  "text/template"
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
  Service     string  `json:"service"`
  CurrentCost float32 `json:"current_cost"`
}

type WaybarOutput struct {
  Text string `json:"text"`
  Tooltip string `json:"tooltip"`
  Alt string `json:"alt"`
  Class string `json:"class"`
}

func main() {
  var waybarPango bool

  flag.BoolVar(&waybarPango, "waybar-pango", false, "Output Waybar compatible JSON with Pango template per service")
  flag.Parse()

  config, err := Cfg()
  if err != nil {
    panic(err)
  }

  waybarPangoTmpl := template.Must(template.New("waybar").Parse(config.WaybarPango))

  services := new(Services)
  services.initVultr(&config)
  services.refreshVultr()

  if waybarPango == false {
    fmt.Print(services.toJSON())
  } else {
    fmt.Print(services.toWaybar(waybarPangoTmpl))
  }

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

  s.Vultr.Status.Service = "Vultr"
  s.Vultr.Status.CurrentCost = account.PendingCharges
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

  waybarOutput.Text = fmt.Sprintf(
    "%s",
    vultr.String(),
  )

  outputJson, _ := JSONMarshal(waybarOutput)
  return string(outputJson)
}

func JSONMarshal(t interface{}) ([]byte, error) {
  buffer := &bytes.Buffer{}
  encoder := json.NewEncoder(buffer)
  encoder.SetEscapeHTML(false)
  err := encoder.Encode(t)
  return buffer.Bytes(), err
}

