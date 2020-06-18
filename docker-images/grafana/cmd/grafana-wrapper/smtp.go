package main

import (
	"fmt"
)

// grafanaConfigSMTP describes Grafana's SMTP configuration - https://grafana.com/docs/grafana/latest/installation/configuration/#smtp
type grafanaConfigSMTP struct {
	Enabled      bool
	Host         string `ini:",omitempty"`
	User         string `ini:",omitempty"`
	Password     string `ini:",omitempty"`
	FromAddress  string `ini:",omitempty"`
	FromName     string `ini:",omitempty"`
	EhloIdentity string `ini:",omitempty"`
	CertFile     string `ini:",omitempty"`
	KeyFile      string `ini:",omitempty"`
	SkipVerify   bool   `ini:",omitempty"`
}

func newGrafanaSMTPConfig(email *siteEmailConfig) *grafanaConfigSMTP {
	if email == nil || email.SMTP == nil {
		return &grafanaConfigSMTP{Enabled: false}
	}
	return &grafanaConfigSMTP{
		Enabled:      true,
		Host:         fmt.Sprintf("%s:%d", email.SMTP.Host, email.SMTP.Port), // host, port are required
		User:         email.SMTP.Username,
		Password:     email.SMTP.Password,
		EhloIdentity: email.SMTP.Domain, // TODO: is this right?

		// Custom values for Sourcegraph
		FromAddress: email.Address,
		FromName:    "Sourcegraph Grafana",
	}
}
