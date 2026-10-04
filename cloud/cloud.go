package cloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

const refreshTimeout = 30 * time.Second

type Cloud struct {
	Config   *lib.Config   `json:"-"`
	Cache    *lib.Cache    `json:"-"`
	Services []lib.Service `json:"services"`
}

type WaybarOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Alt     string `json:"alt"`
	Class   string `json:"class"`
}

type refreshResult struct {
	status      *lib.ServiceStatus
	err         error
	rateLimited bool
}

func New(config *lib.Config, cache *lib.Cache) *Cloud {
	return &Cloud{Config: config, Cache: cache}
}

func (c *Cloud) AddService(id string, name string, client lib.ServiceClient) {
	c.Services = append(c.Services, lib.Service{
		ID:     id,
		Name:   name,
		Client: client,
	})
}

func (c *Cloud) RefreshAll(ctx context.Context, stderr io.Writer) (failed bool) {
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	results := make([]refreshResult, len(c.Services))

	var wg sync.WaitGroup
	for i, service := range c.Services {
		wg.Go(func() {
			results[i] = refresh(ctx, service.Client)
		})
	}
	wg.Wait()

	updated := time.Now().UTC().Truncate(time.Second)

	for i, result := range results {
		service := &c.Services[i]

		if result.err == nil {
			service.Status = result.status
			if c.Cache != nil {
				c.Cache.Put(service.ID, lib.CacheEntry{
					Updated: updated,
					Status:  result.status,
				})
			}
			continue
		}

		if result.rateLimited && c.Cache != nil {
			if entry, ok := c.Cache.Get(service.ID); ok {
				service.Status = entry.Status
				fmt.Fprintf(
					stderr,
					"%s: rate limited, showing value from %s\n",
					service.ID,
					entry.Updated.Local().Format("2006-01-02 15:04"),
				)
				continue
			}
		}

		failed = true
		fmt.Fprintf(stderr, "%s: %v\n", service.ID, result.err)
	}

	if c.Cache != nil {
		if err := c.Cache.Save(); err != nil {
			fmt.Fprintf(stderr, "cache: %v\n", err)
		}
	}

	return failed
}

func refresh(ctx context.Context, client lib.ServiceClient) refreshResult {
	ctx, statusCode := lib.WithStatusCodeRecorder(ctx)

	status, err := client.GetServiceStatus(ctx)
	if err != nil {
		return refreshResult{
			err: err,
			rateLimited: errors.Is(err, lib.ErrRateLimited) ||
				statusCode() == http.StatusTooManyRequests,
		}
	}

	return refreshResult{status: status}
}

func (c *Cloud) JSON() (string, error) {
	out, err := lib.JSONMarshal(c)
	return string(out), err
}

func (c *Cloud) Waybar(pango *template.Template, usage *template.Template) (string, error) {
	var statuses []string

	for _, service := range c.Services {
		if service.Status == nil {
			continue
		}

		var status bytes.Buffer
		if err := pango.Execute(&status, service); err != nil {
			return "", err
		}

		// Only services that meter usage against a quota render the usage
		// template, so the others don't all end up showing 0%.
		if showUsage(service) {
			if err := usage.Execute(&status, service); err != nil {
				return "", err
			}
		}

		statuses = append(statuses, status.String())
	}

	out, err := lib.JSONMarshal(WaybarOutput{
		Text:    strings.Join(statuses, c.Config.Waybar.PangoJoiner),
		Tooltip: fmt.Sprintf("Updated %s", time.Now().Format(time.RFC822)),
		Class:   "cloudcash",
	})
	return string(out), err
}

func (c *Cloud) Text() string {
	var text strings.Builder

	for _, service := range c.Services {
		// Services whose refresh failed have no status to print.
		if service.Status == nil {
			continue
		}

		fmt.Fprintf(&text, "%-20s", service.Name)

		if !service.UsageOnly() {
			fmt.Fprintf(
				&text,
				"%12s  [previous: %12s / balance: %12s]",
				amount(service.Status.CurrentCharges, service.Status.Currency),
				amount(service.Status.PreviousCharges, service.Status.Currency),
				amount(service.Status.AccountBalance, service.Status.Currency),
			)
		}

		if showUsage(service) {
			fmt.Fprintf(
				&text,
				"  [session: %3s%%%s / weekly: %3s%%%s]",
				service.Status.SessionUsage.RoundBank(0),
				countdown(service.Status.SessionResetsAt, service.Status.SessionResetsIn()),
				service.Status.WeeklyUsage.RoundBank(0),
				countdown(service.Status.WeeklyResetsAt, service.Status.WeeklyResetsIn()),
			)
		}

		text.WriteString("\n")
	}

	return text.String()
}

func (c *Cloud) MenuText(t *template.Template) (string, error) {
	var statuses []string

	for _, service := range c.Services {
		if service.Status == nil {
			continue
		}

		var status bytes.Buffer
		if err := t.Execute(&status, service); err != nil {
			return "", err
		}

		statuses = append(statuses, status.String())
	}

	return strings.Join(statuses, c.Config.Menu.Joiner), nil
}

func ParseTemplate(name string, text string) (*template.Template, error) {
	return template.New(name).
		Funcs(template.FuncMap{"duration": duration}).
		Parse(text)
}

func showUsage(service lib.Service) bool {
	return service.UsageOnly() ||
		service.Status.SessionUsage.IsPositive() ||
		service.Status.WeeklyUsage.IsPositive()
}

func amount(value decimal.Decimal, currency string) string {
	if currency == "" {
		return value.String()
	}

	return value.String() + " " + currency
}

func countdown(resetsAt time.Time, seconds int64) string {
	if resetsAt.IsZero() {
		return ""
	}

	return " (" + duration(seconds) + ")"
}

func duration(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	day := 24 * time.Hour

	switch {
	case d >= day:
		return fmt.Sprintf("%dd%dh", d/day, d%day/time.Hour)
	case d >= time.Hour:
		return fmt.Sprintf("%dh%dm", d/time.Hour, d%time.Hour/time.Minute)
	case d >= time.Minute:
		return fmt.Sprintf("%dm", d/time.Minute)
	default:
		return fmt.Sprintf("%ds", d/time.Second)
	}
}
