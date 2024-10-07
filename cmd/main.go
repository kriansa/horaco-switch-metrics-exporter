package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kriansa/switch-exporter/pkg/exporter"
	"github.com/kriansa/switch-exporter/pkg/scraper"
)

func main() {
	bindAddress := os.Getenv("BIND_ADDRESS")
	verbose := os.Getenv("VERBOSE")
	baseUrl := os.Getenv("SWITCH_URL")
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")

	if baseUrl == "" || user == "" || pass == "" {
		fmt.Println("Usage: switch-exporter")
		fmt.Println("Exports metrics from the ZX-SWTGW218AS switch")
		fmt.Println("")
		fmt.Println("Required ENVs:")
		fmt.Println("  - SWITCH_URL - The HTTP endpoint to the switch")
		fmt.Println("  - USER - The username to sign in on the switch")
		fmt.Println("  - PASS - The password to sign in on the switch")
		fmt.Println("")
		fmt.Println("Optional ENVs:")
		fmt.Println("  - VERBOSE - Set the logging to a more verbose level (default: false)")
		fmt.Println("  - BIND_ADDRESS - Define the address to bind the server to (default: :8080)")
		os.Exit(1)
	}

	if bindAddress == "" {
		bindAddress = ":8080"
	}

	configureLogger(verbose == "true")

	routerScraper := scraper.NewScraper(baseUrl, user, pass)
	if err := exporter.StartServer(bindAddress, routerScraper); err != nil {
		slog.Error("unable to start server", "err", err)
		os.Exit(1)
	}
}

func configureLogger(verbose bool) {
	var log_level slog.Level

	if verbose {
		log_level = slog.LevelDebug
	} else {
		log_level = slog.LevelInfo
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: log_level})
	slog.SetDefault(slog.New(handler))
}
