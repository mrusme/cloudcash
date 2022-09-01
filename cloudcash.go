package main

import (
  "flag"
  "encoding/json"
  "fmt"
  "bytes"

  "text/template"
)

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
  services.initAll(&config)
  services.refreshAll()

  if waybarPango == false {
    fmt.Print(services.toJSON())
  } else {
    fmt.Print(services.toWaybar(waybarPangoTmpl))
  }

  return
}

func JSONMarshal(t interface{}) ([]byte, error) {
  buffer := &bytes.Buffer{}
  encoder := json.NewEncoder(buffer)
  encoder.SetEscapeHTML(false)
  err := encoder.Encode(t)
  return buffer.Bytes(), err
}

