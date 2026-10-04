//go:build darwin

package main

import (
	"xn--gckvb8fzb.com/cloudcash/cloud"
	"xn--gckvb8fzb.com/cloudcash/menu"
)

func runMenu(c *cloud.Cloud) int {
	t, err := cloud.ParseTemplate("menu", c.Config.Menu.Template)
	if err != nil {
		return fail(err)
	}

	menu.Run(c, t)
	return 0
}
