package stackimpact

import (
	"time"
)

type ConfigLoader struct {
	agent *Agent
}

func newConfigLoader(agent *Agent) *ConfigLoader {
	cl := &ConfigLoader{
		agent: agent,
	}

	return cl
}

func (cl *ConfigLoader) start() {
	loadDelay := time.NewTimer(2 * time.Second)
	go func() {
		ph := cl.agent.panicHandler()
		defer ph()

		<-loadDelay.C
		cl.load()
	}()

	loadTicker := time.NewTicker(120 * time.Second)
	go func() {
		for {
			ph := cl.agent.panicHandler()
			defer ph()

			select {
			case <-loadTicker.C:
				cl.load()
			}
		}
	}()
}

func (cl *ConfigLoader) load() {
	payload := map[string]interface{}{}
	if config, err := cl.agent.apiRequest.post("config", payload); err == nil {
		// profiling_enabled yes|no
		if profilingDisabled, exists := config["profiling_disabled"]; exists {
			cl.agent.disableProfiling = (profilingDisabled.(string) == "yes")
		} else {
			cl.agent.disableProfiling = false
		}
	} else {
		cl.agent.log("Error loading config from Dashboard")
		cl.agent.error(err)
	}
}
