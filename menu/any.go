// +build !darwin

package menu

import (
  "fmt"
	"github.com/mrusme/cloudcash/cloud"
)

func Run(c *cloud.Cloud) {
  fmt.Println("Menu not available on this platform!")
  return
}

