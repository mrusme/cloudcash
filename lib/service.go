package lib

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var ErrNotConfigured = errors.New("not configured")

var ErrRateLimited = errors.New("rate limited")

type ServiceClient interface {
	GetServiceStatus(ctx context.Context) (*ServiceStatus, error)
}

type Service struct {
	Client ServiceClient  `json:"-"`
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Status *ServiceStatus `json:"status"`
}

func (s Service) UsageOnly() bool {
	client, ok := s.Client.(interface{ UsageOnly() bool })
	return ok && client.UsageOnly()
}

type ServiceStatus struct {
	Currency        string          `json:"currency"`
	AccountBalance  decimal.Decimal `json:"account_balance"`
	CurrentCharges  decimal.Decimal `json:"current_charges"`
	PreviousCharges decimal.Decimal `json:"previous_charges"`
	// Percentages, for services that meter usage against a quota instead of
	// (or in addition to) charging for it. Zero for everyone else.
	SessionUsage decimal.Decimal `json:"session_usage"`
	WeeklyUsage  decimal.Decimal `json:"weekly_usage"`

	SessionResetsAt time.Time `json:"session_resets_at,omitzero"`
	WeeklyResetsAt  time.Time `json:"weekly_resets_at,omitzero"`
}

func (s ServiceStatus) SessionResetsIn() int64 {
	return secondsUntil(s.SessionResetsAt)
}

func (s ServiceStatus) WeeklyResetsIn() int64 {
	return secondsUntil(s.WeeklyResetsAt)
}

func (s ServiceStatus) MarshalJSON() ([]byte, error) {
	type status ServiceStatus

	out := struct {
		status
		SessionResetsIn *int64 `json:"session_resets_in,omitempty"`
		WeeklyResetsIn  *int64 `json:"weekly_resets_in,omitempty"`
	}{status: status(s)}

	if !s.SessionResetsAt.IsZero() {
		sessionResetsIn := s.SessionResetsIn()
		out.SessionResetsIn = &sessionResetsIn
	}
	if !s.WeeklyResetsAt.IsZero() {
		weeklyResetsIn := s.WeeklyResetsIn()
		out.WeeklyResetsIn = &weeklyResetsIn
	}

	return json.Marshal(out)
}

func secondsUntil(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return max(0, int64(time.Until(t).Seconds()))
}
