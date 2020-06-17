package main

import (
	"fmt"
)

type grafanaConfigSMTP struct {
	Enabled      bool
	Host         string
	User         string
	Password     string
	CertFile     string
	KeyFile      string
	SkipVerify   bool
	FromAddress  string
	FromName     string
	EhloIdentity string
}

func newGrafanaSMTPConfig(email *siteEmailConfig) *grafanaConfigSMTP {
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
