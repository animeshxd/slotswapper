package main

import (
	"os"
	"slotswapper/internal/api"
)

func applyRenderCloudConfig(config *api.Config) {
	if config == nil {
		return
	}
	if os.Getenv("RENDER") != "true" {
		return
	}
	publicUrl := os.Getenv("RENDER_EXTERNAL_URL")
	if publicUrl != "" {
		config.AllowedOrigins = append(config.AllowedOrigins, publicUrl)
	}

	port := os.Getenv("PORT")
	if port != "" {
		config.Addr = ":" + port
	}

}
