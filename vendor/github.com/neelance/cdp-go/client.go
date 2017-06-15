package cdp

import (
	"golang.org/x/net/websocket"

	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/accessibility"
	"github.com/neelance/cdp-go/protocol/animation"
	"github.com/neelance/cdp-go/protocol/applicationcache"
	"github.com/neelance/cdp-go/protocol/browser"
	"github.com/neelance/cdp-go/protocol/cachestorage"
	"github.com/neelance/cdp-go/protocol/console"
	"github.com/neelance/cdp-go/protocol/css"
	"github.com/neelance/cdp-go/protocol/database"
	"github.com/neelance/cdp-go/protocol/debugger"
	"github.com/neelance/cdp-go/protocol/deviceorientation"
	"github.com/neelance/cdp-go/protocol/dom"
	"github.com/neelance/cdp-go/protocol/domdebugger"
	"github.com/neelance/cdp-go/protocol/domsnapshot"
	"github.com/neelance/cdp-go/protocol/domstorage"
	"github.com/neelance/cdp-go/protocol/emulation"
	"github.com/neelance/cdp-go/protocol/heapprofiler"
	"github.com/neelance/cdp-go/protocol/indexeddb"
	"github.com/neelance/cdp-go/protocol/input"
	"github.com/neelance/cdp-go/protocol/inspector"
	"github.com/neelance/cdp-go/protocol/io"
	"github.com/neelance/cdp-go/protocol/layertree"
	"github.com/neelance/cdp-go/protocol/log"
	"github.com/neelance/cdp-go/protocol/memory"
	"github.com/neelance/cdp-go/protocol/network"
	"github.com/neelance/cdp-go/protocol/overlay"
	"github.com/neelance/cdp-go/protocol/page"
	"github.com/neelance/cdp-go/protocol/profiler"
	"github.com/neelance/cdp-go/protocol/runtime"
	"github.com/neelance/cdp-go/protocol/schema"
	"github.com/neelance/cdp-go/protocol/security"
	"github.com/neelance/cdp-go/protocol/serviceworker"
	"github.com/neelance/cdp-go/protocol/storage"
	"github.com/neelance/cdp-go/protocol/systeminfo"
	"github.com/neelance/cdp-go/protocol/target"
	"github.com/neelance/cdp-go/protocol/tethering"
	"github.com/neelance/cdp-go/protocol/tracing"
)

type Client struct {
	*rpc.Client

	Accessibility     accessibility.Client
	Animation         animation.Client
	ApplicationCache  applicationcache.Client
	Browser           browser.Client
	CSS               css.Client
	CacheStorage      cachestorage.Client
	Console           console.Client
	DOM               dom.Client
	DOMDebugger       domdebugger.Client
	DOMSnapshot       domsnapshot.Client
	DOMStorage        domstorage.Client
	Database          database.Client
	Debugger          debugger.Client
	DeviceOrientation deviceorientation.Client
	Emulation         emulation.Client
	HeapProfiler      heapprofiler.Client
	IO                io.Client
	IndexedDB         indexeddb.Client
	Input             input.Client
	Inspector         inspector.Client
	LayerTree         layertree.Client
	Log               log.Client
	Memory            memory.Client
	Network           network.Client
	Overlay           overlay.Client
	Page              page.Client
	Profiler          profiler.Client
	Runtime           runtime.Client
	Schema            schema.Client
	Security          security.Client
	ServiceWorker     serviceworker.Client
	Storage           storage.Client
	SystemInfo        systeminfo.Client
	Target            target.Client
	Tethering         tethering.Client
	Tracing           tracing.Client
}

func Dial(url string) *Client {
	conn, err := websocket.Dial(url, "", url)
	if err != nil {
		panic(err)
	}

	cl := rpc.NewClient(conn)
	return &Client{
		Client: cl,

		Accessibility:     accessibility.Client{Client: cl},
		Animation:         animation.Client{Client: cl},
		ApplicationCache:  applicationcache.Client{Client: cl},
		Browser:           browser.Client{Client: cl},
		CSS:               css.Client{Client: cl},
		CacheStorage:      cachestorage.Client{Client: cl},
		Console:           console.Client{Client: cl},
		DOM:               dom.Client{Client: cl},
		DOMDebugger:       domdebugger.Client{Client: cl},
		DOMSnapshot:       domsnapshot.Client{Client: cl},
		DOMStorage:        domstorage.Client{Client: cl},
		Database:          database.Client{Client: cl},
		Debugger:          debugger.Client{Client: cl},
		DeviceOrientation: deviceorientation.Client{Client: cl},
		Emulation:         emulation.Client{Client: cl},
		HeapProfiler:      heapprofiler.Client{Client: cl},
		IO:                io.Client{Client: cl},
		IndexedDB:         indexeddb.Client{Client: cl},
		Input:             input.Client{Client: cl},
		Inspector:         inspector.Client{Client: cl},
		LayerTree:         layertree.Client{Client: cl},
		Log:               log.Client{Client: cl},
		Memory:            memory.Client{Client: cl},
		Network:           network.Client{Client: cl},
		Overlay:           overlay.Client{Client: cl},
		Page:              page.Client{Client: cl},
		Profiler:          profiler.Client{Client: cl},
		Runtime:           runtime.Client{Client: cl},
		Schema:            schema.Client{Client: cl},
		Security:          security.Client{Client: cl},
		ServiceWorker:     serviceworker.Client{Client: cl},
		Storage:           storage.Client{Client: cl},
		SystemInfo:        systeminfo.Client{Client: cl},
		Target:            target.Client{Client: cl},
		Tethering:         tethering.Client{Client: cl},
		Tracing:           tracing.Client{Client: cl},
	}
}
