package main

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func runCollectHosts(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// start first, then will run with Ticker
	go collectHosts(ctx)

	for {
		select {
		case <-ticker.C:
			go collectHosts(ctx)
		case <-ctx.Done():
			logger.Info("collectHosts ctx done")
			return
		}
	}
}

func collectHosts(ctx context.Context) {
	c := parseConfig()

	_, err := os.Stat(c.AlertManager.ConfigPath)
	if os.IsNotExist(err) {
		createAlertManagerYaml(c)
	}

	if !reflect.DeepEqual(c, config) {
		logger.Info("config changed", zap.String("path", c.AlertManager.ConfigPath))
		createAlertManagerYaml(c)

		resp, err := http.Post(c.AlertManager.ReloadURL, "application/json", nil)
		if err != nil {
			logger.Error("Reload AlertManager Error", zap.Error(err))
		}
		if resp.StatusCode == 200 {
			logger.Info("Reload AlertManager Success")
		}
		config = c
	}

	// put each host inside the queue, they will be checked
	for _, host := range c.Hosts {
		hostQueue <- host
		logger.Debug("collectHosts", zap.String("address", host.Address),
			zap.Strings("emails", host.AlertEmails))
	}
}

func createAlertManagerYaml(c Config) {
	// Generate AlertManagerYaml file
	a := genAlertManagerYaml(c)
	data, err := yaml.Marshal(&a)
	if err != nil {
		logger.Fatal("Yaml Marshal Error", zap.Error(err))
	}
	err = os.WriteFile(c.AlertManager.ConfigPath, data, 0644)
	if err != nil {
		logger.Fatal("WriteFile Error", zap.Error(err), zap.String("path", c.AlertManager.ConfigPath))
	}
	logger.Info("WriteFile Done", zap.String("path", c.AlertManager.ConfigPath))
}
