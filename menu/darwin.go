//go:build darwin

package menu

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"text/template"
	"time"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"

	"xn--gckvb8fzb.com/cloudcash/cloud"
)

func init() {
	runtime.LockOSThread()
}

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
				c.RefreshAll(context.Background(), os.Stderr)

				title, err := c.MenuText(t)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					foundation.Dispatch(func() {
						button.SetTitle(title)
					})
				}

				time.Sleep(time.Hour)
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

	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
