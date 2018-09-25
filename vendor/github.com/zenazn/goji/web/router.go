package web

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type method int

const (
	mCONNECT method = 1 << iota
	mDELETE
	mGET
	mHEAD
	mOPTIONS
	mPATCH
	mPOST
	mPUT
	mTRACE
	// We only natively support the methods above, but we pass through other
	// methods. This constant pretty much only exists for the sake of mALL.
	mIDK

	mALL method = mCONNECT | mDELETE | mGET | mHEAD | mOPTIONS | mPATCH |
		mPOST | mPUT | mTRACE | mIDK
)

// The key used to communicate to the NotFound handler what methods would have
// been allowed if they'd been provided.
const ValidMethodsKey = "goji.web.validMethods"

var validMethodsMap = map[string]method{
	"CONNECT": mCONNECT,
	"DELETE":  mDELETE,
	"GET":     mGET,
	"HEAD":    mHEAD,
	"OPTIONS": mOPTIONS,
	"PATCH":   mPATCH,
	"POST":    mPOST,
	"PUT":     mPUT,
	"TRACE":   mTRACE,
}

type route struct {
	prefix  string
	method  method
	pattern Pattern
	handler Handler
}

type router struct {
	lock     sync.Mutex
	routes   []route
	notFound Handler
	machine  *routeMachine
}

type netHTTPWrap struct {
	http.Handler
}

func (h netHTTPWrap) ServeHTTPC(c C, w http.ResponseWriter, r *http.Request) {
	h.Handler.ServeHTTP(w, r)
}

func parseHandler(h interface{}) Handler {
	switch f := h.(type) {
	case Handler:
		return f
	case http.Handler:
		return netHTTPWrap{f}
	case func(c C, w http.ResponseWriter, r *http.Request):
		return HandlerFunc(f)
	case func(w http.ResponseWriter, r *http.Request):
		return netHTTPWrap{http.HandlerFunc(f)}
	default:
		log.Fatalf("Unknown handler type %v. Expected a web.Handler, "+
			"a http.Handler, or a function with signature func(C, "+
			"http.ResponseWriter, *http.Request) or "+
			"func(http.ResponseWriter, *http.Request)", h)
	}
	panic("log.Fatalf does not return")
}

func httpMethod(mname string) method {
	if method, ok := validMethodsMap[mname]; ok {
		return method
	}
	return mIDK
}

func (rt *router) compile() *routeMachine {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	sm := routeMachine{
		sm:     compile(rt.routes),
		routes: rt.routes,
	}
	rt.setMachine(&sm)
	return &sm
}

func (rt *router) route(c *C, w http.ResponseWriter, r *http.Request) {
	rm := rt.getMachine()
	if rm == nil {
		rm = rt.compile()
	}

	methods, route := rm.route(c, w, r)
	if route != nil {
		route.handler.ServeHTTPC(*c, w, r)
		return
	}

	if methods == 0 {
		rt.notFound.ServeHTTPC(*c, w, r)
		return
	}

	var methodsList = make([]string, 0)
	for mname, meth := range validMethodsMap {
		if methods&meth != 0 {
			methodsList = append(methodsList, mname)
		}
	}
	sort.Strings(methodsList)

	if c.Env == nil {
		c.Env = map[interface{}]interface{}{
			ValidMethodsKey: methodsList,
		}
	} else {
		c.Env[ValidMethodsKey] = methodsList
	}
	rt.notFound.ServeHTTPC(*c, w, r)
}

func (rt *router) handleUntyped(p interface{}, m method, h interface{}) {
	rt.handle(ParsePattern(p), m, parseHandler(h))
}

func (rt *router) handle(p Pattern, m method, h Handler) {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	// Calculate the sorted insertion point, because there's no reason to do
	// swapping hijinks if we're already making a copy. We need to use
	// bubble sort because we can only compare adjacent elements.
	pp := p.Prefix()
	var i int
	for i = len(rt.routes); i > 0; i-- {
		rip := rt.routes[i-1].prefix
		if rip <= pp || strings.HasPrefix(rip, pp) {
			break
		}
	}

	newRoutes := make([]route, len(rt.routes)+1)
	copy(newRoutes, rt.routes[:i])
	newRoutes[i] = route{
		prefix:  pp,
		method:  m,
		pattern: p,
		handler: h,
	}
	copy(newRoutes[i+1:], rt.routes[i:])

	rt.setMachine(nil)
	rt.routes = newRoutes
}
