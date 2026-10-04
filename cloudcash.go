package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"xn--gckvb8fzb.com/cloudcash/cloud"
	"xn--gckvb8fzb.com/cloudcash/cloud/aws"
	"xn--gckvb8fzb.com/cloudcash/cloud/claude"
	"xn--gckvb8fzb.com/cloudcash/cloud/codex"
	"xn--gckvb8fzb.com/cloudcash/cloud/digitalocean"
	"xn--gckvb8fzb.com/cloudcash/cloud/github"
	"xn--gckvb8fzb.com/cloudcash/cloud/hetzner"
	"xn--gckvb8fzb.com/cloudcash/cloud/vultr"
	"xn--gckvb8fzb.com/cloudcash/lib"
)

func main() {
	os.Exit(run())
}

func run() int {
	var jsonOut, waybarPango, menuMode bool

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

	config, err := lib.Cfg()
	if err != nil {
		return fail(err)
	}

	c := cloud.New(&config, openCache())

	err = errors.Join(
		register(c, "vultr", "Vultr", vultr.New),
		register(c, "digitalocean", "DigitalOcean", digitalocean.New),
		register(c, "aws", "AWS", aws.New),
		register(c, "github", "GitHub", github.New),
		register(c, "claude", "Claude", claude.New),
		register(c, "codex", "Codex", codex.New),
		register(c, "hetzner", "Hetzner", hetzner.New),
	)
	if err != nil {
		return fail(err)
	}

	ctx := context.Background()

	switch {
	case menuMode || config.Menu.IsDefault:
		return runMenu(c)

	case jsonOut:
		failed := c.RefreshAll(ctx, os.Stderr)

		out, err := c.JSON()
		if err != nil {
			return fail(err)
		}

		fmt.Print(out)
		return exitCode(failed)

	case waybarPango:
		pango, err := cloud.ParseTemplate("waybar", config.Waybar.Pango)
		if err != nil {
			return fail(err)
		}

		usage, err := cloud.ParseTemplate("waybarUsage", config.Waybar.PangoUsage)
		if err != nil {
			return fail(err)
		}

		c.RefreshAll(ctx, os.Stderr)

		out, err := c.Waybar(pango, usage)
		if err != nil {
			return fail(err)
		}

		fmt.Print(out)
		return 0

	default:
		failed := c.RefreshAll(ctx, os.Stderr)
		fmt.Print(c.Text())
		return exitCode(failed)
	}
}

func register[T lib.ServiceClient](
	c *cloud.Cloud,
	id string,
	name string,
	newClient func(*lib.Config) (T, error),
) error {
	client, err := newClient(c.Config)
	switch {
	case errors.Is(err, lib.ErrNotConfigured):
		return nil
	case err != nil:
		return fmt.Errorf("%s: %w", id, err)
	}

	c.AddService(id, name, client)
	return nil
}

func openCache() *lib.Cache {
	cache, err := lib.NewCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache: %v\n", err)
		return nil
	}

	if err := cache.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "cache: %v\n", err)
	}

	return cache
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, err)
	return 1
}

func exitCode(failed bool) int {
	if failed {
		return 1
	}

	return 0
}
