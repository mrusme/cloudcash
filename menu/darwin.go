// +build darwin

package menu

import (
	"time"

	"github.com/caseymrm/menuet"
	"github.com/mrusme/cloudcash/cloud"
)

func update(c *cloud.Cloud) {
  for {
    c.RefreshAll()
    menuet.App().SetMenuState(&menuet.MenuState{
      Title: c.MenuText(),
    })
    time.Sleep(time.Hour)
  }
}

func Run(c *cloud.Cloud) {
  go update(c)
  menuet.App().Label = "com.github.mrusme.cloudcash"
  menuet.App().RunApplication()
}

