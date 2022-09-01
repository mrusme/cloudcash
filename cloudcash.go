package main

import (
	"flag"
	"fmt"

	"text/template"

	"github.com/mrusme/cloudcash/cloud"
	"github.com/mrusme/cloudcash/cloud/digitalocean"
	"github.com/mrusme/cloudcash/cloud/vultr"
	"github.com/mrusme/cloudcash/lib"
)

func main() {
  var waybarPango bool

  flag.BoolVar(&waybarPango, "waybar-pango", false, "Output Waybar compatible JSON with Pango template per service")
  flag.Parse()

  config, err := lib.Cfg()
  if err != nil {
    panic(err)
  }

  waybarPangoTmpl := template.Must(template.New("waybar").Parse(config.WaybarPango))

  c := cloud.New(&config)

  if s, err := vultr.New(&config); err == nil {
    c.AddService("vultr", "Vultr", s)
  }
  if s, err := digitalocean.New(&config); err == nil {
    c.AddService("digitalocean", "DigitalOcean", s)
  }
  c.RefreshAll()

  if waybarPango == false {
    fmt.Print(c.JSON())
  } else {
    fmt.Print(c.Waybar(waybarPangoTmpl))
  }

  return
}

