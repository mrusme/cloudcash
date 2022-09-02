// +build darwin

package menu

import (
  "fmt"
  "time"
  "text/template"

  "github.com/mrusme/cloudcash/cloud"
  "github.com/progrium/macdriver/cocoa"
  "github.com/progrium/macdriver/core"
  "github.com/progrium/macdriver/objc"
)

func Run(c *cloud.Cloud, t *template.Template) {
  cocoa.TerminateAfterWindowsClose = false
  app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
    obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
    obj.Retain()

    go func() {
      for {
        fmt.Println("Updating menu ...")
        core.Dispatch(func() {
          obj.Button().SetTitle(c.MenuText(t))
        })
        fmt.Println("Sleeping ...")
        time.Sleep(time.Hour)
        fmt.Println("Refreshing ...")
        c.RefreshAll()
      }
    }()

    itemQuit := cocoa.NSMenuItem_New()
    itemQuit.SetTitle("Quit")
    itemQuit.SetAction(objc.Sel("terminate:"))

    menu := cocoa.NSMenu_New()
    menu.AddItem(itemQuit)
    obj.SetMenu(menu)
  })
  fmt.Println("Running menu bar widget ..")
  app.ActivateIgnoringOtherApps(true)
  app.Run()
  fmt.Println("Ended menu bar widget")
}

