// +build !darwin

package menu

import (
  "fmt"
  "text/template"
  "xn--gckvb8fzb.com/cloudcash/cloud"
)

func Run(c *cloud.Cloud, t *template.Template) {
  fmt.Println("Menu not available on this platform!")
  return
}

