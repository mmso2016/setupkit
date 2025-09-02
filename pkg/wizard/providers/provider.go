// Package providers implements different wizard flow providers for SetupKit
package providers

import (
	"fmt"
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// InstallMode defines the installation mode
type InstallMode string

const (
	M