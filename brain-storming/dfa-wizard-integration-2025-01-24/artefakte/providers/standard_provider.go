// Package wizard - Standard DFA Provider implementation
package wizard

import (
	"context"
	"fmt"
	"github.com/mmso2016/setupkit/installer/core"
	"github.com/mmso2016/setupkit/pkg/wizard"
)

// StandardProvider provides the built-in DFA configurations
type StandardProvider struct {
	mode      InstallMode
	config    *core.Config
	dfa       *wizard.DFA
	handlers  map[wizard.State]StateHandler
	uiMappings map[wizard.State]UIStateConfig
}

// Standard states used by the built-in provider
const (
	StateWelcome     wizard.State = "welcome"
	StateModeSelect  wizard.State = "mode_select"
	StateLicense     wizard.State = "license"
	StateComponents  wizard.State = "components"
	StateLocation    wizard.State = "location"
	StateReady       wizard.State = "ready"
	StateInstalling  wizard.State = "installing"
	StateComplete    wizard.State = "complete"
	StateError       wizard.State = "error"
	StateRollback    wizard.State = "rollback"
)

// NewStandardProvider creates a new standard provider
func NewStandardProvider(mode InstallMode, config *core.Config) (*StandardProvider, error) {
	sp := &StandardProvider{
		mode:       mode,
		config:     config,
		handlers:   make(map[wizard.State]StateHandler),
		uiMappings: make(map[wizard.State]UIStateConfig),
	}
	
	// Build DFA based on mode
	if err := sp.buildDFA(); err != nil {
		return nil, err
	}
	
	// Register handlers
	sp.registerHandlers()
	
	// Setup UI mappings
	sp.setupUIMappings()
	
	return sp, nil
}

// ... Rest of the implementation ...
