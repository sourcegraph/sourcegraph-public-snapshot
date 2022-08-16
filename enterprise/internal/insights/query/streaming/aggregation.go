package streaming

import (
	"sort"
	"strings"
)

type aggregated struct {
	resultBufferSize int
	smallestResult   *Aggregate
	Results          map[string]int32
	OtherCount       OtherCount
}

type Aggregate struct {
	Label string
	Count int32
}

type OtherCount struct {
	ResultCount int32
	GroupCount  int32
}

func (a *aggregated) Add(label string, count int32) {
	// 1. We have a match in our in-memory map. Update and update the smallest result.
	// 2. We haven't hit the max buffer size. Add to our in-memory map and update the smallest result.
	// 3. We don't have a match but have a better result than our smallest. Update the overflow by ejected smallest.
	// 4. We don't have a match or a better result. Update the overflow by the hit count.
	if _, ok := a.Results[label]; !ok {
		if len(a.Results) < a.resultBufferSize {
			a.Results[label] = count
			a.updateSmallestAggregate()
		} else {
			newResult := &Aggregate{label, count}
			if a.smallestResult.Less(newResult) {
				delete(a.Results, a.smallestResult.Label)
				a.Results[label] = count
				a.updateSmallestAggregate()
				a.updateOtherCount(a.smallestResult.Count, 1)
			} else {
				a.updateOtherCount(count, 1)
			}
		}
	} else {
		a.Results[label] += count
		a.updateSmallestAggregate()
	}
}

// findSmallestAggregate finds the result with the smallest count and returns it.
func (a *aggregated) findSmallestAggregate() *Aggregate {
	var smallestAggregate *Aggregate
	for label, count := range a.Results {
		tempSmallest := &Aggregate{label, count}
		if smallestAggregate == nil {
			smallestAggregate = tempSmallest
			continue
		}
		if tempSmallest.Less(smallestAggregate) {
			smallestAggregate = tempSmallest
		}
	}
	return smallestAggregate
}

func (a *aggregated) updateSmallestAggregate() {
	smallestResult := a.findSmallestAggregate()
	if smallestResult != nil {
		a.smallestResult = smallestResult
	}
}

func (a *aggregated) updateOtherCount(resultCount, groupCount int32) {
	a.OtherCount.ResultCount += resultCount
	a.OtherCount.GroupCount += groupCount
}

func (a aggregated) SortAggregate() []*Aggregate {
	aggregateSlice := make(aggregateSlice, 0, len(a.Results))
	for val, count := range a.Results {
		aggregateSlice = append(aggregateSlice, &Aggregate{val, count})
	}
	sort.Sort(aggregateSlice)

	return aggregateSlice
}

type aggregateSlice []*Aggregate

func (a *Aggregate) Less(b *Aggregate) bool {
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	if a.Count == b.Count {
		// Sort alphabetically if of same count.
		return strings.Compare(a.Label, b.Label) < 0
	}
	return a.Count > b.Count
}

func (as aggregateSlice) Len() int {
	return len(as)
}

func (as aggregateSlice) Less(i, j int) bool {
	return as[i].Less(as[j])
}

func (as aggregateSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}
