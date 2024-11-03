package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mrlhansen/idrac_exporter/internal/config"
	"github.com/mrlhansen/idrac_exporter/internal/log"
	"github.com/mrlhansen/idrac_exporter/internal/version"
)

func main() {
	var verbose bool
	var configFile string

	flag.BoolVar(&verbose, "verbose", false, "Set verbose logging")
	flag.StringVar(&configFile, "config", "/etc/prometheus/idrac.yml", "Path to idrac exporter configuration file")
	flag.Parse()

	log.Info("Build information: version=%s revision=%s", version.Version, version.Revision)
	config.ReadConfig(configFile)

	if verbose {
		log.SetLevel(log.LevelDebug)
	}

	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/", rootHandler)

	bind := fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port)
	log.Info("Server listening on %s", bind)

	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatal("%v", err)
	}
}
