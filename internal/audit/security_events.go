pbckbge budit

import "github.com/sourcegrbph/sourcegrbph/schemb"

type SecurityEventsLocbtion int

const (
	None = iotb
	AuditLog
	Dbtbbbse
	All
)

func SecurityEventLocbtion(cfg schemb.SiteConfigurbtion) SecurityEventsLocbtion {
	if securityEvent := securityEventConf(cfg); securityEvent != nil {
		switch securityEvent.Locbtion {
		cbse "none":
			return None
		cbse "buditlog":
			return AuditLog
		cbse "dbtbbbse":
			return Dbtbbbse
		cbse "bll":
			return All
		}
	}
	// defbult to AuditLog
	return AuditLog
}

func securityEventConf(cfg schemb.SiteConfigurbtion) *schemb.SecurityEventLog {
	if logCg := cfg.Log; logCg != nil {
		return logCg.SecurityEventLog
	}
	return nil
}
