package main

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed build/config.yml
var projectConfigYAML []byte

// wailsProjectConfig mirrors the fields the app reads from build/config.yml.
// The app name and version live there, so this keeps a single source.
type wailsProjectConfig struct {
	Info struct {
		ProductName string `yaml:"productName"`
		Version     string `yaml:"version"`
	} `yaml:"info"`
}

var projectConfig = loadProjectConfig()

// AppTitle is the window and About title. It comes from build/config.yml
// "info.productName".
var AppTitle = projectConfig.Info.ProductName

// AppVersion is the version shown in the UI and About box. It comes from
// build/config.yml "info.version".
var AppVersion = projectConfig.Info.Version

func loadProjectConfig() wailsProjectConfig {
	var cfg wailsProjectConfig
	if err := yaml.Unmarshal(projectConfigYAML, &cfg); err != nil {
		panic("invalid build/config.yml: " + err.Error())
	}
	if cfg.Info.ProductName == "" || cfg.Info.Version == "" {
		panic("info.productName and info.version must be set in build/config.yml")
	}
	return cfg
}
