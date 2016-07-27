// @flow

import React from "react";
import type {Route} from "react-router";
import Component from "sourcegraph/Component";

import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {getRouteName} from "./routePatterns";
import context from "sourcegraph/app/context";

export function withAppdashRouteStateRecording(ChildComponent: Object): Object {
	type Props = {
		routes: Array<Route>;
		route: React.PropTypes.object.isRequired,
	};

	type State = {
		routes: Array<Route>;
		route: Object,
	};

	class WithAppdashRouteStateRecording extends Component {
		constructor(props: Props) {
			super(props);
			this._recordRenderView = null;
			this._hasMounted = false;
		}

		componentDidMount() {
			this._hasMounted = true;
			this._recordRenderView = null;
			this._recordInitialPageLoad();
		}

		componentDidUpdate() {
			// componentDidUpdate is called only once, and immediately after
			// the initial rendering occurs. At this point the DOM has been
			// fully prepared by React, i.e., the component has been rendered.
			if (this._recordRenderView) {
				let end = new Date().getTime();
				let r = this._recordRenderView;
				recordSpan({
					name: `load view ${r.route}`,
					start: r.start,
					end: end,
					metadata: {
						location: window.location.href,
					},
				});
				this._recordRenderView = null;

				// Update the debug display on the page with the time.
				this._updateDebugTimer(r.start, end);
			}
		}

		reconcileState(state: State, props: Props) {
			Object.assign(state, props);
		}

		onStateTransition(prevState: State, nextState: State) {
			// Begin recording the render time directly after the route changes
			// as this signals loading a new "page" or "view" of sorts.
			let nextRoute = getRouteName(nextState.routes);
			if (this._hasMounted && getRouteName(prevState.routes) !== nextRoute) {
				this._recordRenderView = {
					route: nextRoute,
					start: new Date().getTime(),
				};
			}
		}

		// _recordInitialPageLoad records the initial page load, as observed
		// by the most accurate browser metrics 'window.performance.timing', as
		// a separate span.
		//
		// Unlike the view-related rendering below, this gives us insight into
		// e.g. JS bundle download times, DNS lookup times, etc.
		//
		// TODO(slimsag): for finer-grained access consider sending all of the
		// info in performance.timing to Appdash for display (when available).
		// This would narrow down DNS lookup time, DOM load time, redirection
		// time, etc. (right now we just have page load time, inclusive of
		// everything).
		_recordInitialPageLoad() {
			// Not all browsers (e.g., mobile) support this, but most do.
			if (typeof window.performance === "undefined") return;

			// Record the time between when the browser was ready to fetch the
			// document, and when the document.readyState was changed to "complete".
			// i.e., the time it took to load the page.
			const startTime = window.performance.timing.fetchStart;
			const endTime = window.performance.timing.domComplete > 0 ? window.performance.timing.domComplete : new Date().getTime();
			const routeName = getRouteName(this.state.routes);
			recordSpan({
				name: `load page ${routeName || "(unknown route)"}`,
				start: startTime,
				end: endTime,
				metadata: {
					location: window.location.href,
				},
			});

			// Update the debug display on the page with the time.
			this._updateDebugTimer(startTime, endTime);
		}

		// _updateDebugTimer updates the Appdash debug timer in the lower
		// left-hand corner of the page to represent the given duration (unix
		// timestamps).
		_updateDebugTimer(startTime: number, endTime: number) {
			let debug = document.querySelector("body>#debug>a");
			const loadTimeSeconds = (endTime-startTime) / 1000;

			// $FlowHack
			if (debug) debug.text = `${loadTimeSeconds}s`;
		}

		render() {
			return <ChildComponent {...this.props} />;
		}
	}
	return WithAppdashRouteStateRecording;
}

// RecordSpanOptions represents properties for an Appdash span ("operation").
type RecordSpanOptions = {
	// name is the unique name that identifies the type of operation, but does
	// not include the exact contents/details of the operation.
	//
	// For example, a good name is "/search" because it uniquely identifies the
	// operation (searching), whereas "/search?q=foobar" is a bad name because
	// every recorded operation would be considered unique and could not be
	// aggregated together by Appdash at all within the Dashboard.
	name: string;

	// start is the start time of the span ("operation") in milliseconds, since
	// the unix epoch.
	start: number;

	// end is the end time of the span ("operation") in milliseconds, since the
	// unix epoch.
	end: number;

	// metadata is optional metadata that can be sent with the span. It should
	// be only string keys and values.
	metadata: ?Object;
};

// recordSpan records a single span (operation) to Appdash. Any potential error
// that would occur is sent to console.error instead of being thrown. It is
// no-op if Appdash is not enabled (i.e. context.currentSpanID is not present).
export function recordSpan(opts: RecordSpanOptions) {
	if (!context.currentSpanID) {
		return;
	}
	// TODO(slimsag): use opts.metadata
	defaultFetch(`/.api/internal/appdash/record-span?S=${opts.start}&E=${opts.end}&Name=${opts.name}`, {
		method: "POST",
	})
	.then(checkStatus)
	.catch((err) => {
		console.error("appdash:", err);
	});
}
