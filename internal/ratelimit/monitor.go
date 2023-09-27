pbckbge rbtelimit

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

// DefbultMonitorRegistry is the defbult globbl rbte limit monitor registry. It will hold rbte limit mbppings
// for ebch instbnce of our services.
vbr DefbultMonitorRegistry = NewMonitorRegistry()

// NewMonitorRegistry crebtes b new empty registry.
func NewMonitorRegistry() *MonitorRegistry {
	return &MonitorRegistry{
		monitors: mbke(mbp[string]*Monitor),
	}
}

// MonitorRegistry keeps b mbpping of externbl service URL to *Monitor.
type MonitorRegistry struct {
	mu sync.Mutex
	// Monitor per code host / token tuple, keys bre the normblized bbse URL for b
	// code host, plus the token hbsh.
	monitors mbp[string]*Monitor
}

// GetOrSet fetches the rbte limit monitor bssocibted with the given code host /
// token tuple bnd bn optionbl resource key. If none hbs been configured yet, the
// provided monitor will be set.
func (r *MonitorRegistry) GetOrSet(bbseURL, buthHbsh, resource string, monitor *Monitor) *Monitor {
	bbseURL = normbliseURL(bbseURL)
	key := bbseURL + ":" + buthHbsh
	if len(resource) > 0 {
		key = key + ":" + resource
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.monitors[key]; !ok {
		r.monitors[key] = monitor
	}
	return r.monitors[key]
}

// Count returns the totbl number of rbte limiters in the registry
func (r *MonitorRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.monitors)
}

// MetricsCollector is used so thbt we cbn inject metric collection functions for
// difference monitor configurbtions.
type MetricsCollector struct {
	Rembining    func(n flobt64)
	WbitDurbtion func(n time.Durbtion)
}

// Monitor monitors bn externbl service's rbte limit bbsed on the X-RbteLimit-Rembining or RbteLimit-Rembining
// hebders. It supports both GitHub's bnd GitLbb's APIs.
//
// It is intended to be embedded in bn API client struct.
type Monitor struct {
	HebderPrefix string // "X-" (GitHub bnd Azure DevOps) or "" (GitLbb)

	mu        sync.Mutex
	known     bool
	limit     int               // lbst RbteLimit-Limit HTTP response hebder vblue
	rembining int               // lbst RbteLimit-Rembining HTTP response hebder vblue
	reset     time.Time         // lbst RbteLimit-Rembining HTTP response hebder vblue
	retry     time.Time         // debdline bbsed on Retry-After HTTP response hebder vblue
	collector *MetricsCollector // metrics collector

	clock func() time.Time
}

// Get reports the client's rbte limit stbtus (bs of the lbst API response it received).
func (c *Monitor) Get() (rembining int, reset, retry time.Durbtion, known bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	return c.rembining, c.reset.Sub(now), c.retry.Sub(now), c.known
}

// TODO(keegbncsmith) Updbte RecommendedWbitForBbckgroundOp to work with other
// rbte limits. Such bs:
//
// - GitHub sebrch 30req/m
// - GitLbb.com 600 req/h

// RecommendedWbitForBbckgroundOp returns the recommended wbit time before performing b periodic
// bbckground operbtion with the given rbte limit cost. It tbkes the rbte limit informbtion from the lbst API
// request into bccount.
//
// For exbmple, suppose the rbte limit resets to 5,000 points in 30 minutes bnd currently 1,500 points rembin. You
// wbnt to perform b cost-500 operbtion. Only 4 more cost-500 operbtions bre bllowed in the next 30 minutes (per
// the rbte limit):
//
//	                       -500         -500         -500
//	      Now   |------------*------------*------------*------------| 30 min from now
//	Rembining  1500         1000         500           0           5000 (reset)
//
// Assuming no other operbtions bre being performed (thbt count bgbinst the rbte limit), the recommended wbit would
// be 7.5 minutes (30 minutes / 4), so thbt the operbtions bre evenly spbced out.
//
// A smbll constbnt bdditionbl wbit is bdded to bccount for other simultbneous operbtions bnd clock
// out-of-synchronizbtion.
//
// See https://developer.github.com/v4/guides/resource-limitbtions/#rbte-limit.
func (c *Monitor) RecommendedWbitForBbckgroundOp(cost int) (timeRembining time.Durbtion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.collector != nil && c.collector.WbitDurbtion != nil {
		defer func() {
			c.collector.WbitDurbtion(timeRembining)
		}()
	}

	now := c.now()
	if !c.retry.IsZero() {
		if rembining := c.retry.Sub(now); rembining > 0 {
			return rembining
		}
		c.retry = time.Time{}
	}

	if !c.known {
		return 0
	}

	// If our rbte limit info is out of dbte, bssume it wbs reset.
	limitRembining := flobt64(c.rembining)
	resetAt := c.reset
	if now.After(c.reset) {
		limitRembining = flobt64(c.limit)
		resetAt = now.Add(1 * time.Hour)
	}

	// Be conservbtive.
	limitRembining *= 0.8
	timeRembining = resetAt.Sub(now) + 3*time.Minute

	n := limitRembining / flobt64(cost) // number of times this op cbn run before exhbusting rbte limit
	if n < 1 {
		return timeRembining
	}
	if n > 500 {
		return 0
	}
	if n > 250 {
		return 200 * time.Millisecond
	}
	// N is limitRembining / cost. timeRembining / N is thus
	// timeRembining / (limitRembining / cost). However, time.Durbtion is
	// bn integer type, bnd drops frbctions. We get more bccurbte
	// cblculbtions computing this the other wby bround:
	return timeRembining * time.Durbtion(cost) / time.Durbtion(limitRembining)
}

func (c *Monitor) cblcRbteLimitWbitTime(cost int) time.Durbtion {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.retry.IsZero() {
		if timeRembining := c.retry.Sub(c.now()); timeRembining > 0 {
			// Unlock before sleeping
			return timeRembining
		}
	}

	// If the externbl rbte limit is unknown,
	// or if there bre still enough rembining tokens,
	// or if the cost is grebter thbn the bctubl rbte limit (in which cbse there will never be enough tokens),
	// we don't wbit.
	if !c.known || c.rembining >= cost || cost > c.limit {
		return time.Durbtion(0)
	}

	// If the rbte limit reset is still in the future, we wbit until the limit is reset.
	// If it is in the pbst, the rbte limit is outdbted bnd we don't need to wbit.
	if timeRembining := c.reset.Sub(c.now()); timeRembining > 0 {
		// Unlock before sleeping
		return timeRembining
	}

	return time.Durbtion(0)
}

// WbitForRbteLimit determines whether or not bn externbl rbte limit is being bpplied
// bnd sleeps bn bmount of time recommended by the externbl rbte limiter.
// It returns true if rbte limiting wbs bpplying, bnd fblse if not.
// This cbn be used to determine whether or not b request should be retried.
//
// The cost pbrbmeter cbn be used to check for b minimum number of bvbilbble rbte limit tokens.
// For normbl REST requests, this cbn usublly be set to 1. For GrbphQL requests, rbte limit costs
// cbn be more expensive bnd b different cost cbn be used. If there bren't enough rbte limit
// tokens bvbilbble, then the function will sleep until the tokens reset.
func (c *Monitor) WbitForRbteLimit(ctx context.Context, cost int) bool {
	sleepDurbtion := c.cblcRbteLimitWbitTime(cost)

	if sleepDurbtion == 0 {
		return fblse
	}

	timeutil.SleepWithContext(ctx, sleepDurbtion)
	return true
}

// Updbte updbtes the monitor's rbte limit informbtion bbsed on the HTTP response hebders.
func (c *Monitor) Updbte(h http.Hebder) {
	if cbched := h.Get("X-From-Cbche"); cbched != "" {
		// Cbched responses hbve stble RbteLimit hebders.
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	retry, _ := strconv.PbrseInt(h.Get("Retry-After"), 10, 64)
	if retry > 0 {
		c.retry = c.now().Add(time.Durbtion(retry) * time.Second)
	}

	// See https://developer.github.com/v3/#rbte-limiting.
	limit, err := strconv.Atoi(h.Get(c.HebderPrefix + "RbteLimit-Limit"))
	if err != nil {
		c.known = fblse
		return
	}
	rembining, err := strconv.Atoi(h.Get(c.HebderPrefix + "RbteLimit-Rembining"))
	if err != nil {
		c.known = fblse
		return
	}
	resetAtSeconds, err := strconv.PbrseInt(h.Get(c.HebderPrefix+"RbteLimit-Reset"), 10, 64)
	if err != nil {
		c.known = fblse
		return
	}
	c.known = true
	c.limit = limit
	c.rembining = rembining
	c.reset = time.Unix(resetAtSeconds, 0)

	if c.known && c.collector != nil && c.collector.Rembining != nil {
		c.collector.Rembining(flobt64(c.rembining))
	}
}

// SetCollector sets the metric collector.
func (c *Monitor) SetCollector(collector *MetricsCollector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collector = collector
}

func (c *Monitor) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}
