package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Log creates an INFO log statement that will be a part of the audit log.
// The audit log records comply with the following design: an actor takes an action on an entity within a context.
// Refer to Record struct to see details about individual components.
func Log(ctx context.Context, logger log.Logger, record Record) {
	act := actor.FromContext(ctx)

	// internal actors add a lot of noise to the audit log
	siteConfig := conf.SiteConfig()
	if act.Internal && !IsEnabled(siteConfig, InternalTraffic) {
		return
	}

	client := requestclient.FromContext(ctx)
	auditId := uuid.New().String()
	var fields []log.Field

	fields = append(fields, log.Object("audit",
		log.String("auditId", auditId),
		log.String("entity", record.Entity),
		log.Object("actor",
			log.String("actorUID", actorId(act)),
			log.String("ip", ip(client)),
			log.String("X-Forwarded-For", forwardedFor(client)))))
	fields = append(fields, record.Fields...)

	loggerFunc := getLoggerFuncWithSeverity(logger, siteConfig)
	// message string looks like: #{record.Action} (sampling immunity token: #{auditId})
	loggerFunc(fmt.Sprintf("%s (sampling immunity token: %s)", record.Action, auditId), fields...)
}

func actorId(act *actor.Actor) string {
	if act.UID > 0 {
		return act.UIDString()
	}
	if act.AnonymousUID != "" {
		return act.AnonymousUID
	}
	return "unknown"
}

func ip(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.IP
}

func forwardedFor(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.ForwardedFor
}

type Record struct {
	// Entity is the name of the audited entity
	Entity string
	// Action describes the state change relevant to the audit log
	Action string
	// Fields hold any additional context relevant to the Action
	Fields []log.Field
}

type AuditLogSetting = int

const (
	GitserverAccess = iota
	InternalTraffic
	GraphQL
)

// IsEnabled returns the value of the respective setting from the site config (if set).
// Otherwise, it returns the default value for the setting.
func IsEnabled(cfg schema.SiteConfiguration, setting AuditLogSetting) bool {
	if auditCfg := getAuditCfg(cfg); auditCfg != nil {
		switch setting {
		case GitserverAccess:
			return auditCfg.GitserverAccess
		case InternalTraffic:
			return auditCfg.InternalTraffic
		case GraphQL:
			return auditCfg.GraphQL
		}
	}
	// all settings now currently default to 'false', but that's a coincidence, not intention
	return false
}

// getLoggerFuncWithSeverity returns a specific logger function (logger.Info, logger.Warn, etc.), a the severity is configurable.
func getLoggerFuncWithSeverity(logger log.Logger, cfg schema.SiteConfiguration) func(string, ...log.Field) {
	if auditCfg := getAuditCfg(cfg); auditCfg != nil {
		switch auditCfg.SeverityLevel {
		case "DEBUG":
			return logger.Debug
		case "INFO":
			return logger.Info
		case "WARN":
			return logger.Warn
		case "ERROR":
			return logger.Error
		}
	}
	// default to INFO
	return logger.Info
}

func getAuditCfg(cfg schema.SiteConfiguration) *schema.AuditLog {
	if logCg := cfg.Log; logCg != nil {
		return logCg.AuditLog
	}
	return nil
}
