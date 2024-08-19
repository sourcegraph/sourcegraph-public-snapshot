package outputcore

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// Override allows adjusting the log level for specific scopes.
type Override struct {
	// Scope matches any zapcore.Entry with the name Scope or the prefix Scope + ".".
	Scope string
	// Level is the level to log at for zapcore.Entry's that match Scope.
	Level zapcore.Level
}

// newOverrideCore will create a core with the specific overrides. newCore is
// required because we need to adjust the level to allow the overrides to log.
func newOverrideCore(level zapcore.LevelEnabler, overrides []Override, newCore func(zapcore.LevelEnabler) zapcore.Core) zapcore.Core {
	// We need to adjust level we construct NewCore such that if a level is
	// overriden we end up logging it.
	minOverrideLevel := level
	for _, o := range overrides {
		if !minOverrideLevel.Enabled(o.Level) {
			minOverrideLevel = o.Level
		}
	}

	core := newCore(minOverrideLevel)

	// Only use overrideCore if it could have an effect.
	if minOverrideLevel == level {
		return core
	}

	return &overrideCore{
		Core:      core,
		level:     level,
		overrides: overrides,
	}
}

// overrideCore wraps a core to additionally log a message if it matches
// overrides or level.
type overrideCore struct {
	// Core is the wrapped core. Note its level must be lowered such that if a
	// message matches an override it will be logged. Then we gate everything
	// else with level.
	zapcore.Core

	// level is the passed in level before Core was lowered to take into account
	// the overrides.
	level zapcore.LevelEnabler

	overrides []Override
}

// Level returns the level after lowering due to overrides. We need to return
// the lowest possible level since child cores will shortcut based on this
// level.
func (c *overrideCore) Level() zapcore.Level {
	return zapcore.LevelOf(c.Core)
}

func (c *overrideCore) With(fields []zapcore.Field) zapcore.Core {
	return &overrideCore{
		Core:      c.Core.With(fields),
		level:     c.level,
		overrides: c.overrides,
	}
}

func (c *overrideCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.level.Enabled(ent.Level) {
		return c.Core.Check(ent, ce)
	}

	if !c.Core.Enabled(ent.Level) {
		return ce
	}

	for _, o := range c.overrides {
		if !strings.HasPrefix(ent.LoggerName, o.Scope) {
			continue
		}
		// Check that if o.Scope != ent.LoggerName then it is a child logger
		if len(ent.LoggerName) > len(o.Scope) && ent.LoggerName[len(o.Scope)] != '.' {
			continue
		}

		if o.Level.Enabled(ent.Level) {
			return c.Core.Check(ent, ce)
		}
	}

	return ce
}
