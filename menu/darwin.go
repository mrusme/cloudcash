//go:build darwin
// +build darwin

package menu

import (
	"fmt"
	"text/template"
	"time"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"

	"xn--gckvb8fzb.com/cloudcash/cloud"
)

func Run(c *cloud.Cloud, t *template.Template) {
	appkit.TerminateAfterWindowsClose = false

	app := appkit.Application_SharedApplication()

	app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)

	var statusItem appkit.StatusItem

	app.SetDidFinishLaunching(func(notification foundation.Notification) {
		statusBar := appkit.StatusBar_SystemStatusBar()
		statusItem = statusBar.StatusItemWithLength(appkit.VariableStatusItemLength)
		statusItem.Retain()

		button := statusItem.Button()

		go func() {
			for {
				fmt.Println("Updating menu ...")
				foundation.Dispatch(func() {
					button.SetTitle(c.MenuText(t))
				})
				fmt.Println("Sleeping ...")
				time.Sleep(time.Hour)
				fmt.Println("Refreshing ...")
				c.RefreshAll()
			}
		}()

		menu := appkit.NewMenu()

		itemQuit := appkit.NewMenuItemWithAction(
			"Quit",
			objc.Sel("terminate:"),
			"",
		)

		menu.AddItem(itemQuit)
		statusItem.SetMenu(menu)
	})

	fmt.Println("Running menu bar widget ..")
	app.ActivateIgnoringOtherApps(true)
	app.Run()
	fmt.Println("Ended menu bar widget")
}
