package main

import (
	"flag"
	"fmt"

	"text/template"

	"github.com/mrusme/cloudcash/cloud"
	"github.com/mrusme/cloudcash/cloud/aws"
	"github.com/mrusme/cloudcash/cloud/digitalocean"
	"github.com/mrusme/cloudcash/cloud/vultr"
	"github.com/mrusme/cloudcash/lib"
)

func main() {
  var waybarPango bool = false
  var jsonOut     bool = false

  flag.BoolVar(
    &jsonOut,
    "json",
    false,
    "Output JSON",
  )
  flag.BoolVar(
    &waybarPango,
    "waybar-pango",
    false,
    "Output Waybar compatible JSON with Pango template per service",
  )
  flag.Parse()

  config, err := lib.Cfg()
  if err != nil {
    panic(err)
  }

  c := cloud.New(&config)

  if s, err := vultr.New(&config); err == nil {
    c.AddService("vultr", "Vultr", s)
  }
  if s, err := digitalocean.New(&config); err == nil {
    c.AddService("digitalocean", "DigitalOcean", s)
  }
  if s, err := aws.New(&config); err == nil {
    c.AddService("aws", "AWS", s)
  }
  c.RefreshAll()

  if jsonOut == true {
    fmt.Print(c.JSON())
  } else if waybarPango == true {
    fmt.Print(c.Waybar(
      template.Must(template.New("waybar").Parse(config.Waybar.Pango)),
    ))
  } else {
    fmt.Print(c.Text())
  }

  return
}

