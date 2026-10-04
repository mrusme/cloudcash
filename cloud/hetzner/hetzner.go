package hetzner

import (
	"context"
	"net/http"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

// Hetzner bills traffic in decimal terabytes.
var bytesPerTB = decimal.NewFromInt(1000000000000)

type Hetzner struct {
	cfg *lib.Config
	c   *hcloud.Client
}

// month is the billing window the charges are calculated over.
type month struct {
	start time.Time
	now   time.Time
	hours decimal.Decimal
}

func New(config *lib.Config) (*Hetzner, error) {
	apiKey, err := lib.Secret(
		context.Background(),
		config.Service.Hetzner.APIKey,
		config.Service.Hetzner.APIKeyCommand,
	)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, lib.ErrNotConfigured
	}

	s := new(Hetzner)

	s.cfg = config
	s.c = hcloud.NewClient(
		hcloud.WithToken(apiKey),
		hcloud.WithApplication("cloudcash", lib.UserAgent),
		hcloud.WithHTTPClient(lib.NewHTTPClient()),
	)

	return s, nil
}

func (s *Hetzner) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	pricing, err := s.pricing(ctx)
	if err != nil {
		return nil, err
	}

	m := newMonth(time.Now().UTC())

	charges := decimal.Zero

	servers, err := s.servers(ctx, &pricing, m)
	if err != nil {
		return nil, err
	}
	charges = charges.Add(servers)

	balancers, err := s.loadBalancers(ctx, &pricing, m)
	if err != nil {
		return nil, err
	}
	charges = charges.Add(balancers)

	volumes, err := s.volumes(ctx, &pricing, m)
	if err != nil {
		return nil, err
	}
	charges = charges.Add(volumes)

	primaryIPs, err := s.primaryIPs(ctx, &pricing, m)
	if err != nil {
		return nil, err
	}
	charges = charges.Add(primaryIPs)

	floatingIPs, err := s.floatingIPs(ctx, &pricing, m)
	if err != nil {
		return nil, err
	}
	charges = charges.Add(floatingIPs)

	status := new(lib.ServiceStatus)

	status.Currency = pricing.Currency
	status.CurrentCharges = charges.RoundBank(2)

	return status, nil
}

func (s *Hetzner) pricing(ctx context.Context) (schema.Pricing, error) {
	req, err := s.c.NewRequest(ctx, http.MethodGet, "/pricing", nil)
	if err != nil {
		return schema.Pricing{}, err
	}

	var body schema.PricingGetResponse
	if _, err := s.c.Do(req, &body); err != nil {
		return schema.Pricing{}, err
	}

	return body.Pricing, nil
}

func (s *Hetzner) servers(
	ctx context.Context,
	pricing *schema.Pricing,
	m month,
) (decimal.Decimal, error) {
	charges := decimal.Zero

	servers, err := s.c.Server.All(ctx)
	if err != nil {
		return charges, err
	}

	for _, server := range servers {
		if server.ServerType == nil {
			continue
		}

		p := serverPricing(pricing, server.ServerType.Name, locationName(server.Location))
		if p == nil {
			continue
		}

		hours := m.billedHours(server.Created)
		cost := capped(hours, s.amount(p.PriceHourly), s.amount(p.PriceMonthly))

		// Backups are a surcharge on top of the server itself.
		if server.BackupWindow != "" {
			percentage, err := decimal.NewFromString(pricing.ServerBackup.Percentage)
			if err == nil {
				cost = cost.Add(cost.Mul(percentage).Div(decimal.NewFromInt(100)))
			}
		}

		cost = cost.Add(overage(
			server.OutgoingTraffic,
			server.IncludedTraffic,
			s.amount(p.PricePerTBTraffic),
		))

		charges = charges.Add(cost)
	}

	return charges, nil
}

func (s *Hetzner) loadBalancers(
	ctx context.Context,
	pricing *schema.Pricing,
	m month,
) (decimal.Decimal, error) {
	charges := decimal.Zero

	balancers, err := s.c.LoadBalancer.All(ctx)
	if err != nil {
		return charges, err
	}

	for _, balancer := range balancers {
		if balancer.LoadBalancerType == nil {
			continue
		}

		p := loadBalancerPricing(
			pricing,
			balancer.LoadBalancerType.Name,
			locationName(balancer.Location),
		)
		if p == nil {
			continue
		}

		hours := m.billedHours(balancer.Created)
		cost := capped(hours, s.amount(p.PriceHourly), s.amount(p.PriceMonthly))

		cost = cost.Add(overage(
			balancer.OutgoingTraffic,
			balancer.IncludedTraffic,
			s.amount(p.PricePerTBTraffic),
		))

		charges = charges.Add(cost)
	}

	return charges, nil
}

func (s *Hetzner) volumes(
	ctx context.Context,
	pricing *schema.Pricing,
	m month,
) (decimal.Decimal, error) {
	charges := decimal.Zero

	volumes, err := s.c.Volume.All(ctx)
	if err != nil {
		return charges, err
	}

	perGB := s.amount(pricing.Volume.PricePerGBPerMonth)

	for _, volume := range volumes {
		monthly := perGB.Mul(decimal.NewFromInt(int64(volume.Size)))
		charges = charges.Add(m.prorated(volume.Created, monthly))
	}

	return charges, nil
}

func (s *Hetzner) primaryIPs(
	ctx context.Context,
	pricing *schema.Pricing,
	m month,
) (decimal.Decimal, error) {
	charges := decimal.Zero

	ips, err := s.c.PrimaryIP.All(ctx)
	if err != nil {
		return charges, err
	}

	for _, ip := range ips {
		p := primaryIPPricing(pricing, string(ip.Type), locationName(ip.Location))
		if p == nil {
			continue
		}

		hours := m.billedHours(ip.Created)
		charges = charges.Add(capped(
			hours,
			s.amount(p.PriceHourly),
			s.amount(p.PriceMonthly),
		))
	}

	return charges, nil
}

func (s *Hetzner) floatingIPs(
	ctx context.Context,
	pricing *schema.Pricing,
	m month,
) (decimal.Decimal, error) {
	charges := decimal.Zero

	ips, err := s.c.FloatingIP.All(ctx)
	if err != nil {
		return charges, err
	}

	for _, ip := range ips {
		p := floatingIPPricing(pricing, string(ip.Type), locationName(ip.HomeLocation))
		if p == nil {
			continue
		}

		charges = charges.Add(m.prorated(ip.Created, s.amount(*p)))
	}

	return charges, nil
}

// amount reads a price as net, or gross when the configuration asks for it.
func (s *Hetzner) amount(p schema.Price) decimal.Decimal {
	raw := p.Net
	if s.cfg.Service.Hetzner.Gross {
		raw = p.Gross
	}

	amount, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero
	}

	return amount
}

func newMonth(now time.Time) month {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	return month{
		start: start,
		now:   now,
		hours: decimal.NewFromFloat(start.AddDate(0, 1, 0).Sub(start).Hours()),
	}
}

// billedHours is the time a resource has existed for within this month.
func (m month) billedHours(created time.Time) decimal.Decimal {
	from := m.start
	if created.After(from) {
		from = created
	}

	if !m.now.After(from) {
		return decimal.Zero
	}

	return decimal.NewFromFloat(m.now.Sub(from).Hours())
}

// prorated splits a monthly price across the part of the month a resource has
// existed for. Used for resources the price list only quotes monthly.
func (m month) prorated(created time.Time, monthly decimal.Decimal) decimal.Decimal {
	if m.hours.IsZero() {
		return decimal.Zero
	}

	return monthly.Mul(m.billedHours(created)).Div(m.hours)
}

// capped applies the hourly rate, which Hetzner never charges beyond the price
// of a full month.
func capped(hours decimal.Decimal, hourly decimal.Decimal, monthly decimal.Decimal) decimal.Decimal {
	cost := hours.Mul(hourly)

	if monthly.IsPositive() && cost.GreaterThan(monthly) {
		return monthly
	}

	return cost
}

// overage is the cost of outgoing traffic beyond what the resource includes.
func overage(outgoing uint64, included uint64, perTB decimal.Decimal) decimal.Decimal {
	if outgoing <= included {
		return decimal.Zero
	}

	return decimal.NewFromInt(int64(outgoing - included)).
		Div(bytesPerTB).
		Mul(perTB)
}

func locationName(location *hcloud.Location) string {
	if location == nil {
		return ""
	}

	return location.Name
}

func serverPricing(
	pricing *schema.Pricing,
	serverType string,
	location string,
) *schema.PricingServerTypePrice {
	for _, p := range pricing.ServerTypes {
		if p.Name != serverType {
			continue
		}
		for i, price := range p.Prices {
			if price.Location == location {
				return &p.Prices[i]
			}
		}
	}

	return nil
}

func loadBalancerPricing(
	pricing *schema.Pricing,
	balancerType string,
	location string,
) *schema.PricingLoadBalancerTypePrice {
	for _, p := range pricing.LoadBalancerTypes {
		if p.Name != balancerType {
			continue
		}
		for i, price := range p.Prices {
			if price.Location == location {
				return &p.Prices[i]
			}
		}
	}

	return nil
}

func primaryIPPricing(
	pricing *schema.Pricing,
	ipType string,
	location string,
) *schema.PricingPrimaryIPTypePrice {
	for _, p := range pricing.PrimaryIPs {
		if p.Type != ipType {
			continue
		}
		for i, price := range p.Prices {
			if price.Location == location {
				return &p.Prices[i]
			}
		}
	}

	return nil
}

func floatingIPPricing(
	pricing *schema.Pricing,
	ipType string,
	location string,
) *schema.Price {
	for _, p := range pricing.FloatingIPs {
		if p.Type != ipType {
			continue
		}
		for i, price := range p.Prices {
			if price.Location == location {
				return &p.Prices[i].PriceMonthly
			}
		}
	}

	return nil
}
