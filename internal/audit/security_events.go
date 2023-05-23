package audit

import "github.com/sourcegraph/sourcegraph/schema"

type SecurityEventsLocation int

const (
	None = iota
	AuditLog
	Database
	All
)

func SecurityEventLocation(cfg schema.SiteConfiguration) SecurityEventsLocation {
	if securityEvent := securityEventConf(cfg); securityEvent != nil {
		switch securityEvent.Location {
		case "none":
			return None
		case "auditlog":
			return AuditLog
		case "database":
			return Database
		case "all":
			return All
		}
	}
	// default to AuditLog
	return AuditLog
}

func securityEventConf(cfg schema.SiteConfiguration) *schema.SecurityEventLog {
	if logCg := cfg.Log; logCg != nil {
		return logCg.SecurityEventLog
	}
	return nil
}
