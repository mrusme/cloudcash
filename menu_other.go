//go:build !darwin

package main

import (
	"errors"
	"fmt"

	"xn--gckvb8fzb.com/cloudcash/cloud"
)

func runMenu(*cloud.Cloud) int {
	return fail(fmt.Errorf("menu mode is only available on macOS: %w", errors.ErrUnsupported))
}
