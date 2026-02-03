package main

import (
  "runtime"
  "flag"
  "fmt"

  "text/template"

  "xn--gckvb8fzb.com/cloudcash/cloud"
  "xn--gckvb8fzb.com/cloudcash/cloud/aws"
  "xn--gckvb8fzb.com/cloudcash/cloud/digitalocean"
  "xn--gckvb8fzb.com/cloudcash/cloud/github"
  "xn--gckvb8fzb.com/cloudcash/cloud/vultr"
  "xn--gckvb8fzb.com/cloudcash/lib"
  "xn--gckvb8fzb.com/cloudcash/menu"
)

func main() {
  var waybarPango bool = false
  var jsonOut     bool = false
  var menuMode    bool = false

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
  flag.BoolVar(
    &menuMode,
    "menu-mode",
    false,
    "Run as menubar app (only on macOS)",
  )
  flag.Parse()

  if menuMode == true {
    runtime.LockOSThread()
  }

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
  if s, err := github.New(&config); err == nil {
    c.AddService("github", "GitHub", s)
  }

  c.RefreshAll()

  if menuMode == true ||
     config.Menu.IsDefault == true {
    t := template.Must(template.New("menu").Parse(config.Menu.Template))
    menu.Run(c, t)
    return
  } else {
    if jsonOut == true {
      fmt.Print(c.JSON())
    } else if waybarPango == true {
      fmt.Print(c.Waybar(
        template.Must(template.New("waybar").Parse(config.Waybar.Pango)),
      ))
    } else {
      fmt.Print(c.Text())
    }
  }

  return
}

