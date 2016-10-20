package stackimpact

import (
	"math"
	"time"
)

type metricFuncType func() float64
type reportFuncType func(trigger string)

type ReportingStrategy struct {
	agent           *Agent
	delay           int
	interval        int
	metricFunc      metricFuncType
	reportFunc      reportFuncType
	reporting       bool
	anomalyReported bool
	measurements    []float64
}

func newReportingStrategy(agent *Agent, delay int, interval int, metricFunc metricFuncType, reportFunc reportFuncType) *ReportingStrategy {
	rs := &ReportingStrategy{
		agent:           agent,
		delay:           delay,
		interval:        interval,
		metricFunc:      metricFunc,
		reportFunc:      reportFunc,
		reporting:       false,
		anomalyReported: false,
		measurements:    make([]float64, 0),
	}

	return rs
}

func (rs *ReportingStrategy) start() {

	if rs.metricFunc != nil {
		anomalyTicker := time.NewTicker(1 * time.Second)
		go func() {
			ph := rs.agent.panicHandler()
			defer ph()

			for {
				select {
				case <-anomalyTicker.C:
					if rs.checkAnomaly() && !rs.anomalyReported {
						rs.executeReport(TriggerAnomaly)
						rs.anomalyReported = true
					}
				}
			}
		}()
	}

	delayTimer := time.NewTimer(time.Duration(rs.delay) * time.Second)
	go func() {
		ph := rs.agent.panicHandler()
		defer ph()

		<-delayTimer.C
		rs.executeReport(TriggerTimer)
		rs.anomalyReported = false

		intervalTicker := time.NewTicker(time.Duration(rs.interval) * time.Second)
		go func() {
			ph := rs.agent.panicHandler()
			defer ph()

			for {
				select {
				case <-intervalTicker.C:
					rs.executeReport(TriggerTimer)
					rs.anomalyReported = false
				}
			}
		}()
	}()
}

func (rs *ReportingStrategy) checkAnomaly() bool {
	m := rs.metricFunc()
	rs.measurements = append(rs.measurements, m)
	l := len(rs.measurements)

	if l < 60 {
		return false
	} else if l > 60 {
		rs.measurements = rs.measurements[l-60 : l]
	}

	mean, dev := stdev(rs.measurements[0:30])

	recent := rs.measurements[50:60]
	anomalies := 0
	for _, m := range recent {
		if m > mean+2*dev && m > mean+mean*0.5 {
			anomalies++
		}
	}

	return anomalies >= 5
}

func stdev(numbers []float64) (float64, float64) {
	mean := 0.0
	for _, number := range numbers {
		mean += number
	}
	mean = mean / float64(len(numbers))

	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}

	variance := total / float64(len(numbers)-1)

	return mean, math.Sqrt(variance)
}

func (rs *ReportingStrategy) executeReport(trigger string) {
	if !rs.reporting {
		rs.agent.overheadLock.Lock()
		rs.reporting = true
		rs.reportFunc(trigger)
		rs.reporting = false
		rs.agent.overheadLock.Unlock()
	}
}
