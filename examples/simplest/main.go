package main

import "github.com/setupkit"

func main() {
	// The absolute simplest installer - just 5 lines!
	setupkit.Install(setupkit.Config{
		AppName: "Hello App",
		Version: "1.0.0",
	})
}
