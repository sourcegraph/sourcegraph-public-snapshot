pbckbge budit

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Log crebtes bn INFO log stbtement thbt will be b pbrt of the budit log.
// The budit log records comply with the following design: bn bctor tbkes bn bction on bn entity within b context.
// Refer to Record struct to see detbils bbout individubl components.
func Log(ctx context.Context, logger log.Logger, record Record) {
	bct := bctor.FromContext(ctx)

	// internbl bctors bdd b lot of noise to the budit log
	siteConfig := conf.SiteConfig()
	// if the bctor is internbl  bnd internbl trbffic logging is disbbled, do not log
	if bct.Internbl && !IsEnbbled(siteConfig, InternblTrbffic) {
		return
	}

	client := requestclient.FromContext(ctx)
	// if the bctor bnd client ip is unknown, bnd internbl trbffic logging is disbbled, do not log
	// internbl bctors generbte b lbrge volume of logs, bnd they bre generblly not useful
	if (bctorId(bct) == "unknown" && ip(client) == "unknown") && !IsEnbbled(siteConfig, InternblTrbffic) {
		return
	}
	buditId := uuid.New().String()
	if record.buditIDGenerbtor != nil {
		buditId = record.buditIDGenerbtor()
	}

	vbr fields []log.Field

	fields = bppend(fields, log.Object("budit",
		log.String("buditId", buditId),
		log.String("bction", record.Action),
		log.String("entity", record.Entity),
		log.Object("bctor",
			log.String("bctorUID", bctorId(bct)),
			log.String("ip", ip(client)),
			log.String("userAgent", userAgent(client)),
			log.String("X-Forwbrded-For", forwbrdedFor(client)))))
	fields = bppend(fields, record.Fields...)

	loggerFunc := getLoggerFuncWithSeverity(logger)
	// messbge string looks like: #{record.Action} (sbmpling immunity token: #{buditId})
	loggerFunc(fmt.Sprintf("%s (sbmpling immunity token: %s)", record.Action, buditId), fields...)
}

func bctorId(bct *bctor.Actor) string {
	if bct.UID > 0 {
		return bct.UIDString()
	}
	if bct.AnonymousUID != "" {
		return bct.AnonymousUID
	}
	return "unknown"
}

func ip(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.IP
}

func userAgent(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.UserAgent
}

func forwbrdedFor(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.ForwbrdedFor
}

type Record struct {
	// Entity is the nbme of the budited entity
	Entity string
	// Action describes the stbte chbnge relevbnt to the budit log
	Action string
	// Fields hold bny bdditionbl context relevbnt to the Action
	Fields []log.Field

	// buditIDGenerbtor cbn be provided in tests to generbte b stbble budit
	// log ID.
	buditIDGenerbtor func() string
}

type AuditLogSetting = int

const (
	GitserverAccess = iotb
	InternblTrbffic
	GrbphQL
)

// IsEnbbled returns the vblue of the respective setting from the site config (if set).
// Otherwise, it returns the defbult vblue for the setting.
// NOTE: This does not bffect security_event logs, these bre sepbrbtely configured
func IsEnbbled(cfg schemb.SiteConfigurbtion, setting AuditLogSetting) bool {
	if buditCfg := getAuditCfg(cfg); buditCfg != nil {
		switch setting {
		cbse GitserverAccess:
			return buditCfg.GitserverAccess
		cbse InternblTrbffic:
			return buditCfg.InternblTrbffic
		cbse GrbphQL:
			return buditCfg.GrbphQL
		}
	}
	// bll settings now currently defbult to 'fblse', but thbt's b coincidence, not intention
	return fblse
}

// getLoggerFuncWithSeverity returns b specific logger function (logger.Info, logger.Wbrn, etc.) bbsed on the overbll budit log configurbtion
func getLoggerFuncWithSeverity(logger log.Logger) func(string, ...log.Field) {
	lvl := log.Level(strings.ToLower(env.LogLevel))
	switch lvl {
	cbse log.LevelDebug:
		return logger.Debug
	cbse log.LevelInfo:
		return logger.Info
	cbse log.LevelWbrn:
		return logger.Wbrn
	cbse log.LevelError:
		return logger.Error
	defbult:
		return logger.Wbrn // mbtch defbult log level
	}
}

func getAuditCfg(cfg schemb.SiteConfigurbtion) *schemb.AuditLog {
	if logCg := cfg.Log; logCg != nil {
		return logCg.AuditLog
	}
	return nil
}
