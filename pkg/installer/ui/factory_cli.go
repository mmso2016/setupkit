//go:build !nocli
// +build !nocli

package ui

import (
	"github.com/mmso2016/setupkit/pkg/installer/core"
	"github.com/mmso2016/setupkit/pkg/installer/ui/cli"
)

// createCLI creates a DFA-controlled CLI-based UI
func createCLI() (core.UI, error) {
	return cli.NewDFA(), nil
}
