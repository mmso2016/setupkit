package installer_test

import (
	"context"
	"fmt"
	"testing"
	
	"github.com/mmso2016/setupkit/installer"
)

// BenchmarkInstallerCreation benchmarks installer creation
func BenchmarkInstallerCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = installer.New(
			installer.WithAppName("BenchApp"),
			installer.WithVersion("1.0.0"),
			installer.WithInstallDir("/tmp/bench"),
		)
	}
}

// BenchmarkInstallerWithComponents benchmarks installer with multiple components
func BenchmarkInstallerWithComponents(b *testing.B) {
	components := make([]installer.Component, 10)
	for i := range components {
		components[i] = installer.Component{
			ID:          fmt.Sprintf("comp%d", i),
			Name:        fmt.Sprintf("Component %d", i),
			Description: "Benchmark component",
			Installer: func(ctx context.Context) error {
				return nil
			},
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = installer.New(
			installer.WithAppName("BenchApp"),
			installer.WithComponents(components...),
		)
	}
}

// BenchmarkLoggerOperations benchmarks logger operations
func BenchmarkLoggerOperations(b *testing.B) {
	logger := installer.NewLogger("info", "")
	
	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("Benchmark message", "iteration", i)
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Debug("Debug message", "iteration", i)
		}
	})
	
	b.Run("Verbose", func(b *testing.B) {
		logger.SetVerbose(true)
		for i := 0; i < b.N; i++ {
			logger.Verbose("Verbose message", "iteration", i)
		}
	})
}

// BenchmarkErrorCreation benchmarks error creation
func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = installer.NewError(
			installer.ExitGeneralError,
			"Benchmark error",
			nil,
		)
	}
}

// BenchmarkExitCodeLookup benchmarks exit code description lookup
func BenchmarkExitCodeLookup(b *testing.B) {
	codes := []int{
		installer.ExitSuccess,
		installer.ExitGeneralError,
		installer.ExitPermissionError,
		installer.ExitUserCancelled,
		999, // Unknown code
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = installer.ExitCodeDescription(codes[i%len(codes)])
	}
}

// BenchmarkComponentSelection benchmarks component selection
func BenchmarkComponentSelection(b *testing.B) {
	components := make([]installer.Component, 20)
	for i := range components {
		components[i] = installer.Component{
			ID:       fmt.Sprintf("comp%d", i),
			Name:     fmt.Sprintf("Component %d", i),
			Selected: i%2 == 0,
			Required: i%5 == 0,
		}
	}
	
	inst, _ := installer.New(
		installer.WithAppName("BenchApp"),
		installer.WithComponents(components...),
	)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selected := make([]installer.Component, 0, 10)
		for j := range components {
			if j%3 == 0 {
				selected = append(selected, components[j])
			}
		}
		inst.SetSelectedComponents(selected)
	}
}

// BenchmarkPathConfiguration benchmarks PATH configuration
func BenchmarkPathConfiguration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pathConfig := &installer.PathConfiguration{
			Enabled: true,
			System:  i%2 == 0,
			Dirs:    []string{"/bin", "/sbin", "/usr/bin", "/usr/local/bin"},
		}
		
		_, _ = installer.New(
			installer.WithAppName("BenchApp"),
			installer.WithPathConfig(pathConfig),
		)
	}
}
