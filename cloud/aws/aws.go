package aws

import (
  "errors"
  "time"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/costexplorer"
  "github.com/shopspring/decimal"

  "xn--gckvb8fzb.com/cloudcash/lib"
)

type AWS struct {
  c          *costexplorer.CostExplorer
}

func New(config *lib.Config) (*AWS, error) {
  if config.Service.AWS.AWSAccessKeyID == "" ||
     config.Service.AWS.AWSSecretAccessKey == "" ||
     config.Service.AWS.Region == "" {
    return nil, errors.New("No API key")
  }

  s := new(AWS)

  sess, err := session.NewSession(&aws.Config{
    Region: aws.String(config.Service.AWS.Region),
    Credentials: credentials.NewStaticCredentials(config.Service.AWS.AWSAccessKeyID, config.Service.AWS.AWSSecretAccessKey, ""),
  })
  if err != nil {
    return nil, err
  }

  s.c = costexplorer.New(sess)

  return s, nil
}

func (s *AWS) GetServiceStatus() (*lib.ServiceStatus, error) {
  start := time.Now().Format("2006-01-02")
  end := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
  granularity := "MONTHLY"
  metrics := []string{
    "UnblendedCost",
    "UsageQuantity",
  }

  result, err := s.c.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
    TimePeriod: &costexplorer.DateInterval{
      Start: aws.String(start),
      End:   aws.String(end),
    },
    Granularity: aws.String(granularity),
    GroupBy: []*costexplorer.GroupDefinition{
      &costexplorer.GroupDefinition{
        Type: aws.String("DIMENSION"),
        Key:  aws.String("SERVICE"),
      },
    },
    Metrics: aws.StringSlice(metrics),
  })
  if err != nil {
    return nil, err
  }

  status := new(lib.ServiceStatus)

  status.AccountBalance, _ = decimal.NewFromString("0.0")
  if len(result.ResultsByTime) > 0 {
    if unblendedCost, ok := result.ResultsByTime[0].Total["UnblendedCost"]; ok {
      status.CurrentCharges, _ = decimal.NewFromString(*unblendedCost.Amount)
    }
  }
  status.PreviousCharges, _ = decimal.NewFromString("0.0")

  return status, nil
}


