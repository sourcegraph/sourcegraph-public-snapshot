// +build js

package main

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

func init() {
	js.Global.Set("MustScrollTo", MustScrollTo)

	processHashSet := func() {
		// Scroll to hash target.
		targetID := strings.TrimPrefix(dom.GetWindow().Location().Hash, "#")
		target, ok := document.GetElementByID(targetID).(dom.HTMLElement)
		if ok {
			centerWindowOn(target)
		}

		processHash(target)
	}
	// Jump to desired hash after page finishes loading (and override browser's default hash jumping).
	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		go func() {
			// This needs to be in a goroutine or else it "happens too early".
			// TODO: See if there's a better event than DOMContentLoaded.
			processHashSet()
		}()
	})
	// Start watching for hashchange events.
	dom.GetWindow().AddEventListener("hashchange", false, func(event dom.Event) {
		event.PreventDefault()

		processHashSet()
	})

	document.AddEventListener("keydown", false, func(event dom.Event) {
		if event.DefaultPrevented() {
			return
		}
		// Ignore when some element other than body has focus (it means the user is typing elsewhere).
		if !event.Target().IsEqualNode(document.Body()) {
			return
		}

		switch ke := event.(*dom.KeyboardEvent); {
		// Escape.
		case ke.KeyCode == 27 && !ke.Repeat && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey:
			// TODO: dom.GetWindow().History().ReplaceState(...)
			js.Global.Get("window").Get("history").Call("replaceState", nil, nil, "#")

			processHashSet()

			ke.PreventDefault()
		}
	})
}

// targetID must point to a valid target.
func MustScrollTo(targetID string) {
	target := document.GetElementByID(targetID).(dom.HTMLElement)

	// TODO: dom.GetWindow().History().ReplaceState(...)
	js.Global.Get("window").Get("history").Call("replaceState", nil, nil, "#"+targetID)

	// TODO: Decide if it's better to do this or not to.
	//centerWindowOn(target)

	processHash(target)
}

// processHash highlights the selected element by giving it a "hash-selected" class.
// target can be nil if there isn't a valid target.
func processHash(target dom.HTMLElement) {
	// Clear everything.
	for _, e := range document.GetElementsByClassName("hash-selected") {
		e.Class().Remove("hash-selected")
	}

	if target != nil {
		target.Class().Add("hash-selected")
	}
}

// centerWindowOn scrolls window so that (the middle of) target is in the middle of window.
func centerWindowOn(target dom.HTMLElement) {
	windowHalfHeight := dom.GetWindow().InnerHeight() / 2
	targetHalfHeight := target.OffsetHeight() / 2
	if targetHalfHeight > float64(windowHalfHeight)*0.8 { // Prevent top of target from being offscreen.
		targetHalfHeight = float64(windowHalfHeight) * 0.8
	}
	dom.GetWindow().ScrollTo(dom.GetWindow().ScrollX(), int(offsetTopRoot(target)+targetHalfHeight)-windowHalfHeight)
}

// offsetTopRoot returns the offset top of element e relative to root element.
func offsetTopRoot(e dom.HTMLElement) float64 {
	var offsetTopRoot float64
	for ; e != nil; e = e.OffsetParent() {
		offsetTopRoot += e.OffsetTop()
	}
	return offsetTopRoot
}
