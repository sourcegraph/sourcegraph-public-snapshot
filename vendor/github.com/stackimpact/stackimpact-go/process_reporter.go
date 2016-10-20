package stackimpact

import (
	"runtime"
	"syscall"
	"time"
)

type ProcessReporter struct {
	agent   *Agent
	metrics map[string]*Metric
}

func newProcessReporter(agent *Agent) *ProcessReporter {
	pr := &ProcessReporter{
		agent:   agent,
		metrics: make(map[string]*Metric),
	}

	return pr
}

func (pr *ProcessReporter) start() {
	go pr.report()

	collectTicker := time.NewTicker(1 * time.Minute)

	go func() {
		ph := pr.agent.panicHandler()
		defer ph()

		for {
			select {
			case <-collectTicker.C:
				pr.report()
			}
		}
	}()
}

func (pr *ProcessReporter) reportMetric(typ string, category string, name string, unit string, value float64) *Metric {
	key := typ + category + name
	var metric *Metric
	if existingMetric, exists := pr.metrics[key]; !exists {
		metric = newMetric(pr.agent, typ, category, name, unit)
		pr.metrics[key] = metric
	} else {
		metric = existingMetric
	}

	metric.createMeasurement(TriggerTimer, value, nil)

	if metric.hasMeasurement() {
		pr.agent.messageQueue.addMessage("metric", metric.toMap())
	}

	return metric
}

func (pr *ProcessReporter) report() {
	cpuTime, err := readCPUTime()
	if err == nil {
		cpuTimeMetric := pr.reportMetric(TypeCounter, CategoryCPU, NameCPUTime, UnitNanosecond, float64(cpuTime))
		if cpuTimeMetric.hasMeasurement() {
			cpuUsage := (float64(cpuTimeMetric.measurement.value) / float64(60*1e9)) * 100
			pr.reportMetric(TypeState, CategoryCPU, NameCPUUsage, UnitPercent, float64(cpuUsage))
		}
	} else {
		pr.agent.error(err)
	}

	maxRSS, err := readMaxRSS()
	if err == nil {
		pr.reportMetric(TypeState, CategoryMemory, NameMaxRSS, UnitKilobyte, float64(maxRSS))
	} else {
		pr.agent.error(err)
	}

	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	pr.reportMetric(TypeState, CategoryMemory, NameAllocated, UnitByte, float64(memStats.Alloc))
	pr.reportMetric(TypeCounter, CategoryMemory, NameLookups, UnitNone, float64(memStats.Lookups))
	pr.reportMetric(TypeCounter, CategoryMemory, NameMallocs, UnitNone, float64(memStats.Mallocs))
	pr.reportMetric(TypeCounter, CategoryMemory, NameFrees, UnitNone, float64(memStats.Frees))
	pr.reportMetric(TypeState, CategoryMemory, NameHeapObjects, UnitNone, float64(memStats.HeapObjects))
	pr.reportMetric(TypeCounter, CategoryGC, NameGCTotalPause, UnitNanosecond, float64(memStats.PauseTotalNs))
	pr.reportMetric(TypeCounter, CategoryGC, NameNumGC, UnitNone, float64(memStats.NumGC))
	pr.reportMetric(TypeState, CategoryGC, NameGCCPUFraction, UnitNone, float64(memStats.GCCPUFraction))

	numGoroutine := runtime.NumGoroutine()
	pr.reportMetric(TypeState, CategoryRuntime, NameNumGoroutines, UnitNone, float64(numGoroutine))
}

func readCPUTime() (int64, error) {
	rusage := new(syscall.Rusage)
	if err := syscall.Getrusage(0, rusage); err != nil {
		return 0, err
	}

	var cpuTimeNanos int64
	cpuTimeNanos =
		int64(rusage.Utime.Sec*1e9) +
			int64(rusage.Utime.Usec) +
			int64(rusage.Stime.Sec*1e9) +
			int64(rusage.Stime.Usec)

	return cpuTimeNanos, nil
}

func readMaxRSS() (int64, error) {
	rusage := new(syscall.Rusage)
	if err := syscall.Getrusage(0, rusage); err != nil {
		return 0, err
	}

	var maxRSS int64
	maxRSS = int64(rusage.Maxrss)

	if runtime.GOOS == "darwin" {
		maxRSS = maxRSS / 1000
	}

	return maxRSS, nil
}
