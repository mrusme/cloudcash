package lib

import (
  "github.com/shopspring/decimal"
)

type ServiceClient interface {
  GetServiceStatus() (*ServiceStatus, error)
}

type Service struct {
  Client     ServiceClient  `json:"-"`
  ID         string         `json:"id"`
  Name       string         `json:"name"`
  Status     *ServiceStatus `json:"status"`
}

type ServiceStatus struct {
  AccountBalance  decimal.Decimal `json:"account_balance"`
  CurrentCharges  decimal.Decimal `json:"current_charges"`
  PreviousCharges decimal.Decimal `json:"previous_charges"`
  // Percentages, for services that meter usage against a quota instead of
  // (or in addition to) charging for it. Zero for everyone else.
  SessionUsage    decimal.Decimal `json:"session_usage"`
  WeeklyUsage     decimal.Decimal `json:"weekly_usage"`
}

