// +build !darwin

package menu

import (
  "fmt"
  "text/template"
  "github.com/mrusme/cloudcash/cloud"
)

func Run(c *cloud.Cloud, t *template.Template) {
  fmt.Println("Menu not available on this platform!")
  return
}

