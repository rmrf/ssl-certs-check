package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type Host struct {
	Address     string   `toml:"address"`
	AlertEmails []string `toml:"alert-emails"`
}
type Config struct {
	ListenAddress         string       `toml:"listen-address"`
	RefreshIntervalSecond int          `toml:"refresh-interval-second"`
	Concurrency           int          `toml:"concurrency"`
	AlertManager          AlertManager `toml:"alertmanager"`
	Hosts                 []Host       `toml:"hosts"`
}
type AlertManager struct {
	ReloadURL     string `toml:"reload-url"`
	ConfigPath    string `toml:"config-path"`
	SMTPSmarthost string `toml:"smtp-smarthost"`
	SMTPFrom      string `toml:"smtp-from"`
	SMTPUsername  string `toml:"smtp-username"`
	SMTPPassword  string `toml:"smtp-password"`
}

type AlertManagerYaml struct {
	Global struct {
		SMTPSmarthost    string `yaml:"smtp_smarthost"`
		SMTPFrom         string `yaml:"smtp_from"`
		SMTPAuthUsername string `yaml:"smtp_auth_username"`
		SMTPAuthPassword string `yaml:"smtp_auth_password"`
		ResolveTimeout   string `yaml:"resolve_timeout"`
	} `yaml:"global"`
	Templates []string `yaml:"templates"`
	Route     struct {
		GroupBy        []string `yaml:"group_by"`
		GroupWait      string   `yaml:"group_wait"`
		GroupInterval  string   `yaml:"group_interval"`
		RepeatInterval string   `yaml:"repeat_interval"`
		Receiver       string   `yaml:"receiver"`
		Routes         []struct {
			Matchers []string `yaml:"matchers"`
			Receiver string   `yaml:"receiver"`
		} `yaml:"routes"`
	} `yaml:"route"`
	Receivers []struct {
		Name         string `yaml:"name"`
		EmailConfigs []struct {
			To string `yaml:"to"`
		} `yaml:"email_configs,omitempty"`
	} `yaml:"receivers"`
}

func parseConfig() Config {
	var c Config
	_, err := toml.DecodeFile(*configFile, &c)
	if err != nil {
		logger.Error("Could not Decode toml configuration file", zap.Error(err))
		os.Exit(1)
	}
	return c
}

func genAlertManagerYaml(c Config) AlertManagerYaml {
	var emailCount = make(map[string]int)

	var a AlertManagerYaml
	a.Global.ResolveTimeout = "1m"
	a.Global.SMTPAuthPassword = c.AlertManager.SMTPPassword
	a.Global.SMTPAuthUsername = c.AlertManager.SMTPUsername
	a.Global.SMTPFrom = c.AlertManager.SMTPFrom
	a.Global.SMTPSmarthost = c.AlertManager.SMTPSmarthost

	a.Templates = []string{"/etc/alertmanager/template/*.tmpl"}

	a.Route.GroupBy = []string{"alert_email"}
	a.Route.GroupWait = "30s"
	a.Route.GroupInterval = "5m"
	a.Route.RepeatInterval = "24h"
	a.Route.Receiver = "default"

	for _, host := range c.Hosts {
		for _, e := range host.AlertEmails {
			emailCount[e]++
		}
	}

	for email := range emailCount {
		a.Route.Routes = append(a.Route.Routes, struct {
			Matchers []string `yaml:"matchers"`
			Receiver string   `yaml:"receiver"`
		}{
			Matchers: []string{"alert_email=" + email},
			Receiver: email,
		})

		a.Receivers = append(a.Receivers, struct {
			Name         string `yaml:"name"`
			EmailConfigs []struct {
				To string `yaml:"to"`
			} `yaml:"email_configs,omitempty"`
		}{
			Name: email,
			EmailConfigs: []struct {
				To string `yaml:"to"`
			}{
				{
					To: email,
				},
			},
		})

	}
	a.Receivers = append(a.Receivers, struct {
		Name         string `yaml:"name"`
		EmailConfigs []struct {
			To string `yaml:"to"`
		} `yaml:"email_configs,omitempty"`
	}{Name: "default"})

	return a
}
