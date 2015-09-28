package appdash

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

func init() {
	RegisterEvent(AggregateEvent{})
}

// AggregateEvent represents an aggregated set of timespan events. This is the
// only type of event produced by the AggregateStore type.
type AggregateEvent struct {
	// The root span name of every item in this aggregated set of timespan events.
	Name string

	// Trace IDs for the slowest of the above times (useful for inspection).
	Slowest []ID
}

// Schema implements the Event interface.
func (e AggregateEvent) Schema() string { return "aggregate" }

// TODO(slimsag): do not encode aggregate events in JSON. We have to do this for
// now because the reflection code can't handle *Trace types sufficiently.

// MarshalEvent implements the EventMarshaler interface.
func (e AggregateEvent) MarshalEvent() (Annotations, error) {
	// Encode the entire event as JSON.
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return Annotations{
		{Key: "JSON", Value: data},
	}, nil
}

// UnmarshalEvent implements the EventUnmarshaler interface.
func (e AggregateEvent) UnmarshalEvent(as Annotations) (Event, error) {
	// Find the annotation with our key.
	for _, ann := range as {
		if ann.Key != "JSON" {
			continue
		}
		err := json.Unmarshal(ann.Value, &e)
		if err != nil {
			return nil, fmt.Errorf("AggregateEvent.UnmarshalEvent: %v", err)
		}
		return e, nil
	}
	return nil, errors.New("expected one annotation with key=\"JSON\"")
}

// spanGroupSlowest represents one of the slowest traces in a span group.
type spanGroupSlowest struct {
	TraceID    ID        // Root span ID of the slowest trace.
	Start, End time.Time // Start and end time of the slowest trace.
}

// empty tells if this spanGroupSlowest slot is empty / uninitialized.
func (s spanGroupSlowest) empty() bool {
	return s == spanGroupSlowest{}
}

// spanGroup represents all of the times for the root spans (i.e. traces) of the
// given name. It also contains the N-slowest traces of the group.
type spanGroup struct {
	// Trace is the trace ID that the generated AggregateEvent has been placed
	// into for collection.
	Trace     SpanID
	Name      string             // Root span name (e.g. the route for httptrace).
	Times     [][2]time.Time     // Aggregated timespans for the traces.
	TimeSpans []ID               // SpanID.Span of each associated TimespanEvent for the Times slice
	Slowest   []spanGroupSlowest // N-slowest traces in the group.
}

func (s spanGroup) Len() int      { return len(s.Slowest) }
func (s spanGroup) Swap(i, j int) { s.Slowest[i], s.Slowest[j] = s.Slowest[j], s.Slowest[i] }
func (s spanGroup) Less(i, j int) bool {
	a := s.Slowest[i]
	b := s.Slowest[j]

	// A sorts before B if it took a greater amount of time than B (slowest
	// to-fastest sorting).
	return a.End.Sub(a.Start) > b.End.Sub(b.Start)
}

// update updates the span group to account for a potentially slowest trace,
// returning whether or not the given trace was indeed slowest. The timespan ID
// is the SpanID.Span of the TimespanEvent for future removal upon eviction.
func (s *spanGroup) update(start, end time.Time, timespan ID, trace ID, remove func(trace ID)) bool {
	s.Times = append(s.Times, [2]time.Time{start, end})
	s.TimeSpans = append(s.TimeSpans, timespan)

	// The s.Slowest list is kept sorted from slowest to fastest. As we want to
	// steal the slot from the fastest (or zero) one we iterate over it
	// backwards comparing times.
	for i := len(s.Slowest) - 1; i > 0; i-- {
		sm := s.Slowest[i]
		if sm.TraceID == trace {
			// Trace is already inside the group as one of the slowest.
			return false
		}

		// If our time is lesser than the trace in the slot already, we aren't
		// slower so don't steal the slot.
		if end.Sub(start) < sm.End.Sub(sm.Start) {
			continue
		}

		// If there is already a trace inside this group (i.e. we are taking its
		// spot as one of the slowest), then we must request for its removal from
		// the output store.
		if sm.TraceID != 0 {
			remove(sm.TraceID)
		}

		s.Slowest[i] = spanGroupSlowest{
			TraceID: trace,
			Start:   start,
			End:     end,
		}
		sort.Sort(s)
		return true
	}
	return false
}

// evictBefore evicts all times in the group
func (s *spanGroup) evictBefore(tnano int64, debug bool, deleteSub func(s SpanID)) {
	count := 0
search:
	for i, ts := range s.Times {
		if ts[0].UnixNano() < tnano {
			s.Times = append(s.Times[:i], s.Times[i+1:]...)

			// Remove the associated subspan holding the TimespanEvent in the
			// output MemoryStore.
			id := s.TimeSpans[i]
			s.TimeSpans = append(s.TimeSpans[:i], s.TimeSpans[i+1:]...)
			deleteSub(SpanID{Trace: s.Trace.Trace, Span: id, Parent: s.Trace.Span})

			count++
			goto search
		}
	}

	if debug && count > 0 {
		log.Printf("AggregateStore: evicted %d timespans from the group %q", count, s.Name)
	}
}

// The AggregateStore collection process can be described as follows:
//
// 1. Collection on AggregateStore occurs.
// 3. Collection is sent directly to pre-storage
//   - i.e. LimitStore backed by its own MemoryStore.
// 4. Eviction runs if needed.
//   - Every group has an eviction process ran; removes times older than 72/hrs.
//   - Each N-slowest trace in the group older than 72hr is evicted from output.
//   - Empty span groups (no trace over past 72/hr) are removed entirely.
// 5. Find a group for the collection
//   - Only succeeds if a spanName has or is being been collected.
//   - Otherwise collections end up in pre-storage until we get the spanName.
// 6. Collection is unmarshaled into a set of events, trace time is determined.
// 7. Group is updated to consider the collection as being one of the N-slowest.
//   - Older N-slowest trace is removed.
// 8. N-slowest trace collections that are in pre-storage:
//   - Removed from pre-storage.
//   - Placed into output MemoryStore.
// 9. Data Storage
//   - Aggregation data is stored as a phony trace (so same storage backends can be used).
//   - The old AggregationEvent is removed from output MemoryStore.
//   - The new AggregationEvent with updated N-slowest trace IDs is inserted.
//   - A TimespanEvent (subspan) is recorded into the trace.
//     - Not stored in AggregationEvent as a slice (because O(N) vs O(1) performance for updates).

// AggregateStore aggregates timespan events into groups based on the root span
// name. Much like a RecentStore, it evicts aggregated events after a certain
// time period.
type AggregateStore struct {
	// MinEvictAge is the minimum age of group data before it is evicted.
	MinEvictAge time.Duration

	// MaxRate is the maximum expected rate of incoming traces.
	//
	// Multiple traces can be collected at once, and they must be queued up in
	// memory until the entire trace has been collected, otherwise the N-slowest
	// traces cannot be stored.
	//
	// If the number is too large, a lot of memory will be used (to store
	// MaxRate number of traces), and if too small some then some aggregate
	// events will not have the full N-slowest traces associated with them.
	MaxRate int

	// NSlowest is the number of slowest traces to fully keep for inspection for
	// each group.
	NSlowest int

	// Debug is whether or not to log debug messages.
	Debug bool

	// MemoryStore is the memory store were aggregated traces are saved to and
	// deleted from. It is the final destination for traces.
	*MemoryStore

	mu           sync.Mutex
	groups       map[ID]*spanGroup // map of trace ID to span group.
	insertTimes  map[ID]time.Time  // map of times that groups was inserted into at
	groupsByName map[string]ID     // looks up a groups trace ID by name.
	pre          *LimitStore       // traces which do not have span groups yet
	lastEvicted  time.Time         // last time that eviction ran
}

// NewAggregateStore is short-hand for:
//
//  store := &AggregateStore{
//      MinEvictAge: 72 * time.Hour,
//      MaxRate: 4096,
//      NSlowest: 5,
//      MemoryStore: NewMemoryStore(),
//  }
//
func NewAggregateStore() *AggregateStore {
	return &AggregateStore{
		MinEvictAge: 72 * time.Hour,
		MaxRate:     4096,
		NSlowest:    5,
		MemoryStore: NewMemoryStore(),
	}
}

// Collect calls the underlying store's Collect, deleting the oldest
// trace if the capacity has been reached.
func (as *AggregateStore) Collect(id SpanID, anns ...Annotation) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Initialization
	if as.groups == nil {
		as.groups = make(map[ID]*spanGroup)
		as.insertTimes = make(map[ID]time.Time)
		as.groupsByName = make(map[string]ID)
		as.pre = &LimitStore{
			Max:         as.MaxRate,
			DeleteStore: NewMemoryStore(),
		}
		go as.clearGroups()
	}

	if as.Debug {
		// Determine the total number of traces and times in each named span
		// group.
		nTraces := 0
		nTimes := 0
		for _, id := range as.groupsByName {
			g, ok := as.groups[id]
			if !ok {
				continue
			}
			for _, sm := range g.Slowest {
				if sm.TraceID != 0 {
					nTraces++
				}
			}
			nTimes += len(g.Times)
		}

		// Log some statistics: these can be used to identify serious issues
		// relating to overstorage or memory leakage in the primary data maps.
		msTraces, err := as.MemoryStore.Traces()
		if err != nil {
			log.Println(err)
		}
		exceeding := len(msTraces) - (len(as.groupsByName) * as.NSlowest)
		if exceeding < 0 {
			exceeding = 0
		}
		nextEvict := as.MinEvictAge - time.Since(as.lastEvicted)
		log.Printf("AggregateStore: [%d groups by ID] [%d groups by name] [%d-slowest traces] [%d trace times]\n", len(as.groups), len(as.groupsByName), nTraces, nTimes)
		log.Printf("AggregateStore: [%d traces in MemoryStore; exceeding us by %d] [eviction in %s]\n", len(msTraces), exceeding, nextEvict)

		// Validate that the N-slowest traces we store are not exceeding what the user asked for.
		if nTraces > 0 && (len(as.groupsByName)/nTraces) > as.NSlowest {
			log.Println("AggregateStore: WARNING: Have too many N-slowest traces for each span group:")
			for _, id := range as.groupsByName {
				g := as.groups[id]
				log.Printf("AggregateStore: %q has %d-slowest traces\n", g.Name, len(g.Slowest))
			}
		}
	}

	// Collect into the limit store.
	if err := as.pre.Collect(id, anns...); err != nil {
		return err
	}

	// Consider eviction of old data.
	if time.Since(as.lastEvicted) > as.MinEvictAge {
		// Function for evictBefore to invoke when removing TimespanEvents that
		// we've previously stored in the output MemoryStore.
		deleteSub := func(id SpanID) {
			as.MemoryStore.Lock()
			if !as.MemoryStore.deleteSubNoLock(id, false) {
				panic("failed to delete spanID")
			}
			as.MemoryStore.Unlock()
		}
		if err := as.evictBefore(time.Now().Add(-1*as.MinEvictAge), deleteSub); err != nil {
			return err
		}
	}

	// Grab the group for our span.
	group, ok := as.group(id, anns...)
	if !ok {
		// We don't have a group for the trace, and can't create one (the
		// spanName event isn't present yet).
		return nil
	}

	// Unmarshal the events.
	var events []Event
	if err := UnmarshalEvents(anns, &events); err != nil {
		return err
	}

	// Find the start and end time of the trace.
	eStart, eEnd, ok := findTraceTimes(events)
	if !ok {
		// We didn't find any timespan events at all, so we're done here.
		return nil
	}

	// Update the group to consider this trace being one of the slowest.
	timespanID := NewSpanID(group.Trace)
	group.update(eStart, eEnd, timespanID.Span, id.Trace, func(trace ID) {
		// Delete the request trace from the output store.
		if err := as.deleteOutput(trace); err != nil {
			log.Printf("AggregateStore: failed to delete a trace: %s", err)
		}
	})

	// Move traces from the limit store into the group, as needed.
	for _, slowest := range group.Slowest {
		// Find the trace in the limit store.
		trace, err := as.pre.Trace(slowest.TraceID)
		if err == ErrTraceNotFound {
			continue
		}
		if err != nil {
			return err
		}

		// Place into output store.
		var walk func(t *Trace) error
		walk = func(t *Trace) error {
			err := as.MemoryStore.Collect(t.Span.ID, t.Span.Annotations...)
			if err != nil {
				return err
			}
			for _, sub := range t.Sub {
				if err := walk(sub); err != nil {
					return err
				}
			}
			return nil
		}
		if err := walk(trace); err != nil {
			return err
		}

		// Delete from the limit store.
		err = as.pre.Delete(slowest.TraceID)
		if err != nil {
			return err
		}
	}

	// Prepare the aggregation event (before locking below).
	ev := &AggregateEvent{
		Name: group.Name,
	}
	for _, slowest := range group.Slowest {
		if !slowest.empty() {
			ev.Slowest = append(ev.Slowest, slowest.TraceID)
		}
	}
	if as.Debug && len(ev.Slowest) == 0 {
		log.Printf("AggregateStore: no slowest traces for group %q (consider increasing MaxRate)", group.Name)
	}

	// Prepare the timespan event (also before locking below).
	tev := &timespanEvent{
		S: eStart,
		E: eEnd,
	}

	// As we're updating the aggregation event, we go ahead and delete the old
	// one now. We do this all under as.MemoryStore.Lock otherwise users (e.g. the
	// web UI) can pull from as.MemoryStore when the trace has been deleted.
	as.MemoryStore.Lock()
	defer as.MemoryStore.Unlock()
	as.MemoryStore.deleteSubNoLock(group.Trace, true)

	// Record an aggregate event with the given name.
	recEvent := func(e Event, spanID SpanID) error {
		anns, err := MarshalEvent(e)
		if err != nil {
			return err
		}
		return as.MemoryStore.collectNoLock(spanID, anns...)
	}
	if err := recEvent(spanName{Name: group.Name}, group.Trace); err != nil {
		return err
	}
	if err := recEvent(ev, group.Trace); err != nil {
		return err
	}

	// Record the timespan event as a subspan of the aggregation event.
	if err := recEvent(tev, timespanID); err != nil {
		return err
	}
	return nil
}

// deleteOutput deletes the given traces from the output memory store.
func (as *AggregateStore) deleteOutput(traces ...ID) error {
	for _, trace := range traces {
		if err := as.MemoryStore.Delete(trace); err != nil {
			return err
		}
	}
	return nil
}

// group returns the span group that the collection belongs in, or nil, false if
// no such span group exists / could be created.
//
// The as.mu lock must be held for this method to operate safely.
func (as *AggregateStore) group(id SpanID, anns ...Annotation) (*spanGroup, bool) {
	// Do nothing if we already have a group associated with our root span.
	if group, ok := as.groups[id.Trace]; ok {
		return group, true
	}

	// At this point, we need a root span or else we can't create the group.
	if !id.IsRoot() {
		return nil, false
	}

	// And likewise, always a name event.
	var name spanName
	if err := UnmarshalEvent(anns, &name); err != nil {
		return nil, false
	}

	// If there already exists a group with that name, then we just associate
	// our trace with that group for future lookup and we're good to go.
	if groupID, ok := as.groupsByName[name.Name]; ok {
		group := as.groups[groupID]
		as.groups[id.Trace] = group
		as.insertTimes[id.Trace] = time.Now()
		return group, true
	}

	// Create a new group, and associate our trace with it.
	group := &spanGroup{
		Name:    name.Name,
		Trace:   NewRootSpanID(),
		Slowest: make([]spanGroupSlowest, as.NSlowest),
	}
	as.groups[id.Trace] = group
	as.insertTimes[id.Trace] = time.Now()
	as.groupsByName[name.Name] = id.Trace
	return group, true
}

// clearGroups removes IDs from as.groups once they are old enough to no longer
// need to be alive (i.e. after we're certain no more collections will occur for
// that ID). It is used so the map does not leak memory.
//
// TODO(slimsag): find a more correct solution to this. Maybe we can get rid of
// as.groups all-together and have no need for clearing them here?
func (as *AggregateStore) clearGroups() {
	deleteAfter := 30 * time.Second
	for {
		time.Sleep(deleteAfter)

		as.mu.Lock()
	removal:
		for id, _ := range as.groups {
			if time.Since(as.insertTimes[id]) > deleteAfter {
				for _, nameID := range as.groupsByName {
					if id == nameID {
						continue removal
					}
				}
				delete(as.insertTimes, id)
				delete(as.groups, id)
			}
		}
		as.mu.Unlock()
	}
}

// evictBefore evicts aggregation events that were created before t.
//
// The as.mu lock must be held for this method to operate safely.
func (as *AggregateStore) evictBefore(t time.Time, deleteSub func(id SpanID)) error {
	evictStart := time.Now()
	as.lastEvicted = evictStart
	tnano := t.UnixNano()

	// Build a list of aggregation events to evict.
	var toEvict []ID
	for _, group := range as.groups {
		group.evictBefore(tnano, as.Debug, deleteSub)

	searchSlowest:
		for i, sm := range group.Slowest {
			if !sm.empty() && sm.Start.UnixNano() < tnano {
				group.Slowest = append(group.Slowest[:i], group.Slowest[i+1:]...)
				toEvict = append(toEvict, sm.TraceID)
				goto searchSlowest
			}
		}

		// If the group is not completely empty, we have nothing more to do.
		if len(group.Times) > 0 || len(group.Slowest) > 0 {
			continue
		}

		// Remove the empty group from the maps, noting that as.groups often
		// has multiple references to the same group.
		for id, g := range as.groups {
			if g == group {
				delete(as.groups, id)
			}
		}
		delete(as.groupsByName, group.Name)

		// Also request removal of the group (AggregateEvent) from the output store.
		err := as.deleteOutput(group.Trace.Trace)
		if err != nil {
			return err
		}
	}

	// We are done if there is nothing to evict.
	if len(toEvict) == 0 {
		return nil
	}

	if as.Debug {
		log.Printf("AggregateStore: deleting %d slowest traces created before %s (age check took %s)", len(toEvict), t, time.Since(evictStart))
	}

	// Spawn separate goroutine so we don't hold the as.mu lock.
	go func() {
		deleteStart := time.Now()
		if err := as.deleteOutput(toEvict...); err != nil {
			log.Printf("AggregateStore: failed to delete slowest traces: %s", err)
		}
		if as.Debug {
			log.Printf("AggregateStore: finished deleting %d slowest traces created before %s (took %s)", len(toEvict), t, time.Since(deleteStart))
		}
	}()
	return nil
}

// findTraceTimes finds the minimum and maximum timespan event times for the
// given set of events, or returns ok == false if there are no such events.
func findTraceTimes(events []Event) (start, end time.Time, ok bool) {
	// Find the start and end time of the trace.
	var (
		eStart, eEnd time.Time
		haveTimes    = false
	)
	for _, e := range events {
		e, ok := e.(TimespanEvent)
		if !ok {
			continue
		}
		if !haveTimes {
			haveTimes = true
			eStart = e.Start()
			eEnd = e.End()
			continue
		}
		if v := e.Start(); v.UnixNano() < eStart.UnixNano() {
			eStart = v
		}
		if v := e.End(); v.UnixNano() > eEnd.UnixNano() {
			eEnd = v
		}
	}
	if !haveTimes {
		// We didn't find any timespan events at all, so we're done here.
		ok = false
		return
	}
	return eStart, eEnd, true
}
