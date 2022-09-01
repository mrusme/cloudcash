package main

import (
  "context"
  "os"
  "encoding/json"
  "fmt"

  "github.com/vultr/govultr/v2"
  "golang.org/x/oauth2"
)

type Output struct {
  Vultr struct  {
    CurrentCost float32 `json:"current_cost"`
  } `json:"vultr,omitempty"`
}

func main() {
  ctx := context.Background()
  oauth2cfg := oauth2.Config{}
  ts := oauth2cfg.TokenSource(ctx, &oauth2.Token{AccessToken: os.Getenv("VULTR_API_KEY")})
  vultr := govultr.NewClient(oauth2.NewClient(ctx, ts))
  vultr.SetUserAgent("github.com/mrusme/cloudcash")

  account, err := vultr.Account.Get(ctx)
  if err != nil {
    return
  }

  output := Output{}
  output.Vultr.CurrentCost = account.PendingCharges

  outputJson, _ := json.Marshal(output)
  fmt.Println(string(outputJson))

  return
}
