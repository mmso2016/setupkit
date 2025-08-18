//go:build !nocli
// +build !nocli

package ui

import (
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/installer/ui/cli"
)

// createCLI creates a CLI-based UI
func createCLI() (core.UI, error) {
	return cli.New(), nil
}
