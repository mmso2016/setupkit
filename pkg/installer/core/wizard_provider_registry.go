// Package core - Auto-registration of standard wizard providers
package core

func init() {
	// Auto-register standard wizard providers
	RegisterWizardProvider("standard-express", NewStandardWizardProvider(ModeExpress))
	RegisterWizardProvider("standard-custom", NewStandardWizardProvider(ModeCustom))
	RegisterWizardProvider("standard-advanced", NewStandardWizardProvider(ModeAdvanced))
}