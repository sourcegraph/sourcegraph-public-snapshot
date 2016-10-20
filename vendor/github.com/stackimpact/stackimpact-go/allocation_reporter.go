package stackimpact

import (
	"fmt"
	"math"
	"runtime"
	"sort"
)

type recordSorter []runtime.MemProfileRecord

func (x recordSorter) Len() int {
	return len(x)
}

func (x recordSorter) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x recordSorter) Less(i, j int) bool {
	return x[i].InUseBytes() > x[j].InUseBytes()
}

func readMemAlloc() float64 {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return float64(memStats.Alloc)
}

type AllocationReporter struct {
	agent             *Agent
	reportingStrategy *ReportingStrategy
}

func newAllocationReporter(agent *Agent) *AllocationReporter {
	ar := &AllocationReporter{
		agent:             agent,
		reportingStrategy: nil,
	}

	ar.reportingStrategy = newReportingStrategy(agent, 100, 300,
		func() float64 {
			memAlloc := readMemAlloc()
			return float64(int64(memAlloc / 1e6))
		},
		func(trigger string) {
			ar.agent.log("Allocation report triggered by reporting strategy, trigger=%v", trigger)
			ar.report(trigger)
		},
	)

	return ar
}

func (ar *AllocationReporter) start() {
	ar.reportingStrategy.start()
}

func (ar *AllocationReporter) report(trigger string) {
	if ar.agent.disableProfiling {
		return
	}

	records, err := ar.readMemoryProfile()
	if err != nil {
		ar.agent.error(err)
		return
	}

	// allocated size
	if callGraph, err := ar.createAllocationCallGraph(records); err != nil {
		ar.agent.error(err)
	} else {
		// filter calls with lower than 10KB
		callGraph.filter(10000, math.Inf(0))

		metric := newMetric(ar.agent, TypeProfile, CategoryMemoryProfile, NameAllocatedSize, UnitByte)
		metric.createMeasurement(trigger, callGraph.measurement, callGraph)
		ar.agent.messageQueue.addMessage("metric", metric.toMap())
	}
}

func (ar *AllocationReporter) readMemoryProfile() ([]runtime.MemProfileRecord, error) {
	var records []runtime.MemProfileRecord
	n, ok := runtime.MemProfile(nil, false)
	for {
		records = make([]runtime.MemProfileRecord, n+50)
		n, ok = runtime.MemProfile(records, false)
		if ok {
			records = records[0:n]
			break
		}
	}

	return records, nil
}

func (ar *AllocationReporter) createAllocationCallGraph(records []runtime.MemProfileRecord) (*BreakdownNode, error) {
	if len(records) > 100 {
		sort.Sort(recordSorter(records))
		records = records[0:100]
	}

	rootNode := newBreakdownNode("root")

	for i := range records {
		record := &records[i]

		var measurement float64
		measurement = float64(record.InUseBytes())

		if err := addStackToCallGraph(rootNode, record.Stack(), measurement); err != nil {
			return nil, err
		}
	}

	return rootNode, nil
}

func addStackToCallGraph(rootNode *BreakdownNode, stk []uintptr, measurement float64) error {
	frames := make([]*BreakdownNode, 0)

	// create stack frames
	wasPanic := false
	for i, pc := range stk {
		f := runtime.FuncForPC(pc)
		if f == nil {
			wasPanic = false
		} else {
			tracepc := pc

			// Back up to call instruction.
			if i > 0 && pc > f.Entry() && !wasPanic {
				if runtime.GOARCH == "386" || runtime.GOARCH == "amd64" {
					tracepc--
				} else {
					tracepc -= 4 // arm, etc
				}
			}

			funcName := f.Name()
			fileName, fileLine := f.FileLine(tracepc)

			if funcName == "runtime.goexit" {
				continue
			}

			frameName := fmt.Sprintf("%v (%v:%v)", funcName, fileName, fileLine)
			frame := newBreakdownNode(frameName)
			frames = append(frames, frame)

			wasPanic = funcName == "runtime.gopanic"
		}
	}

	// add frames to root
	rootNode.measurement += measurement

	parentNode := rootNode

	for i := len(frames) - 1; i >= 0; i-- {
		f := frames[i]

		child := parentNode.findOrAddChild(f.name)
		child.measurement += measurement

		parentNode = child
	}

	return nil
}
