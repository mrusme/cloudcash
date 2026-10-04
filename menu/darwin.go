//go:build darwin

package menu

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"text/template"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"

	"xn--gckvb8fzb.com/cloudcash/cloud"
)

func init() {
	runtime.LockOSThread()
}

func Run(c *cloud.Cloud, t *template.Template) {
	app := appkit.Application_SharedApplication()

	app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)

	var statusItem appkit.StatusItem

	delegate := &appkit.ApplicationDelegate{}
	delegate.SetApplicationDidFinishLaunching(func(notification foundation.Notification) {
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
					dispatch.MainQueue().DispatchAsync(func() {
						button.SetTitle(title)
					})
				}

				time.Sleep(time.Hour)
			}
		}()

		menu := appkit.NewMenu()

		itemQuit := appkit.NewMenuItemWithSelector(
			"Quit",
			"",
			objc.Sel("terminate:"),
		)

		menu.AddItem(itemQuit)
		statusItem.SetMenu(menu)
	})
	app.SetDelegate(delegate)

	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
