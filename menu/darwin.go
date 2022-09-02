// +build darwin

package menu

import (
  "time"
  "text/template"

  "github.com/caseymrm/menuet"
  "github.com/mrusme/cloudcash/cloud"
)

func update(c *cloud.Cloud, t *template.Template) {
  for {
    c.RefreshAll()
    menuet.App().SetMenuState(&menuet.MenuState{
      Title: c.MenuText(t),
    })
    time.Sleep(time.Hour)
  }
}

func Run(c *cloud.Cloud, t *template.Template) {
  go update(c, t)
  menuet.App().Label = "com.github.mrusme.cloudcash"
  menuet.App().RunApplication()
}

