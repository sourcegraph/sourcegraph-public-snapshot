// tslint:disable: typedef ordered-imports

import "core-js/shim";
import * as React from "react";
import * as ReactDOM from "react-dom";
import "sourcegraph/util/actionLogger";
import "sourcegraph/app/appdash";
import {Router, browserHistory as history, match, applyRouterMiddleware} from "react-router";
import {useScroll} from "react-router-scroll";
import {rootRoute} from "sourcegraph/app/App";
import * as context from "sourcegraph/app/context";
import {resetOnAuthChange} from "sourcegraph/app/resetOnAuthChange";
import {shouldUpdateScroll, hashLinkScroll} from "sourcegraph/app/routerScrollBehavior";
import {AppContainer} from "react-hot-loader";
import Redbox from "redbox-react";

// REQUIRED. Configures Sentry error monitoring.
import "sourcegraph/init/Sentry";

// REQUIRED. Enables HTML history API (pushState) tracking in Google Analytics.
// See https://github.com/googleanalytics/autotrack#shouldtrackurlchange.
import "autotrack/lib/plugins/url-change-tracker";

context.reset(global.window.__sourcegraphJSContext);
resetOnAuthChange();
global.__webpack_public_path__ = document.head.dataset["webpackPublicPath"]; // eslint-disable-line no-undef

const rootEl = document.getElementById("main") as HTMLElement;

let hotReloadCounter = 0;

// matchWithRedirectHandling calls the router match func. If the router issues
// a redirect, it calls match recursively after replacing the location with the
// new one.
function matchWithRedirectHandling(recursed) {
	match({history, routes: rootRoute}, (err, redirectLocation, renderProps) => {
		if (typeof err === "undefined" && typeof redirectLocation === "undefined" && typeof renderProps === "undefined") {
			console.error("404 not found (no route)");
			return;
		}

		if (redirectLocation) {
			let prevLocation = window.location.href;
			history.replace(redirectLocation);
			if (recursed) {
				console.error(`Possible redirect loop: ${prevLocation} -> ${redirectLocation.pathname}`);
				window.location.reload();
				return;
			}
			matchWithRedirectHandling(true);
			return;
		}

		setTimeout(() => {
			ReactDOM.render(
				<AppContainer errorReporter={Redbox}>
					<Router
						key={hotReloadCounter}
						onUpdate={hashLinkScroll}
						{...renderProps as any}
						render={applyRouterMiddleware(useScroll(shouldUpdateScroll))} />
				</AppContainer>,
				rootEl,
			);
		});
	});
}

matchWithRedirectHandling(false);

if (typeof global.module !== "undefined" && global.module.hot) {
	global.module.hot.accept("sourcegraph/app/App", () => {
		hotReloadCounter++;
		matchWithRedirectHandling(false);
	});
}
