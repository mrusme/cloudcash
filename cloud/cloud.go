package cloud

import (
  "bytes"
  "strings"

  "text/template"
  "fmt"
  "time"

  "github.com/mrusme/cloudcash/lib"
)

type Cloud struct {
  Config       *lib.Config   `json:"-"`
  Services     []lib.Service `json:"services"`
}

type WaybarOutput struct {
  Text    string `json:"text"`
  Tooltip string `json:"tooltip"`
  Alt     string `json:"alt"`
  Class   string `json:"class"`
}

func New(config *lib.Config) (*Cloud) {
  c := new(Cloud)
  c.Config = config
  return c
}

func (c *Cloud) AddService(id string, name string, client lib.ServiceClient) (error) {
  c.Services = append(c.Services, lib.Service{
    ID: id,
    Name: name,
    Client: client,
  })

  return nil
}

func (c *Cloud) RefreshAll() () {
  for i := 0; i < len(c.Services); i++ {
    status, err := c.Services[i].Client.GetServiceStatus()
    if err == nil {
      c.Services[i].Status = status
    }
  }
  return
}

func (c *Cloud) JSON() (string) {
  outputJson, _ := lib.JSONMarshal(c)
  return string(outputJson)
}

func (c *Cloud) Waybar(t *template.Template) (string) {
  waybarOutput := new(WaybarOutput)

  var statuses []string

  for _, service := range c.Services {
    var status bytes.Buffer
    if err := t.Execute(&status, service); err == nil {
      statuses = append(statuses, status.String())
    }
  }

  waybarOutput.Class = "cloudcash"
  waybarOutput.Text = strings.Join(statuses, c.Config.Waybar.PangoJoiner)
  waybarOutput.Tooltip = fmt.Sprintf(
    "Updated %s",
    time.Now().Format(time.RFC822),
  )

  outputJson, _ := lib.JSONMarshal(waybarOutput)
  return string(outputJson)
}

func (c *Cloud) Text() (string) {
  var text string = ""

  for _, service := range c.Services {
    text = fmt.Sprintf(
      "%s%-20s$%8s  [previous: $%8s / balance: $%8s]\n",
      text,
      service.Name,
      service.Status.CurrentCharges,
      service.Status.PreviousCharges,
      service.Status.AccountBalance,
    )
  }

  return text
}

func (c *Cloud) MenuText() (string) {
  var text string = ""

  for _, service := range c.Services {
    text = fmt.Sprintf(
      "%s%s %s [$%s/$%s] · ",
      text,
      service.Name,
      service.Status.CurrentCharges,
      service.Status.PreviousCharges,
      service.Status.AccountBalance,
    )
  }

  return text
}
