package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/shopspring/decimal"

	"xn--gckvb8fzb.com/cloudcash/lib"
)

const metric = "UnblendedCost"

type AWS struct {
	c *costexplorer.Client
}

func New(config *lib.Config) (*AWS, error) {
	if config.Service.AWS.AWSAccessKeyID == "" ||
		config.Service.AWS.Region == "" {
		return nil, lib.ErrNotConfigured
	}

	secretAccessKey, err := lib.Secret(
		context.Background(),
		config.Service.AWS.AWSSecretAccessKey,
		config.Service.AWS.AWSSecretAccessKeyCommand,
	)
	if err != nil {
		return nil, err
	}
	if secretAccessKey == "" {
		return nil, lib.ErrNotConfigured
	}

	s := new(AWS)

	s.c = costexplorer.NewFromConfig(aws.Config{
		Region: config.Service.AWS.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			config.Service.AWS.AWSAccessKeyID,
			secretAccessKey,
			"",
		),
		HTTPClient: lib.NewHTTPClient(),
	})

	return s, nil
}

func (s *AWS) GetServiceStatus(ctx context.Context) (*lib.ServiceStatus, error) {
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	previousMonth := currentMonth.AddDate(0, -1, 0)

	result, err := s.c.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(previousMonth.Format(time.DateOnly)),
			End:   aws.String(now.AddDate(0, 0, 1).Format(time.DateOnly)),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{metric},
	})
	if err != nil {
		var limitExceeded *types.LimitExceededException
		if errors.As(err, &limitExceeded) {
			return nil, fmt.Errorf("%w: %w", lib.ErrRateLimited, err)
		}
		return nil, err
	}

	status := new(lib.ServiceStatus)

	for _, period := range result.ResultsByTime {
		cost, ok := period.Total[metric]
		if !ok || cost.Amount == nil || period.TimePeriod == nil {
			continue
		}

		amount, err := decimal.NewFromString(*cost.Amount)
		if err != nil {
			return nil, err
		}

		switch aws.ToString(period.TimePeriod.Start) {
		case currentMonth.Format(time.DateOnly):
			status.CurrentCharges = amount.RoundBank(2)
		case previousMonth.Format(time.DateOnly):
			status.PreviousCharges = amount.RoundBank(2)
		}

		status.Currency = aws.ToString(cost.Unit)
	}

	return status, nil
}
