import "core-js/shim";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { AppContainer } from "react-hot-loader";
import * as Relay from "react-relay";
import { useScroll } from "react-router-scroll";
import Redbox from "redbox-react";

import { Router, applyRouterMiddleware, browserHistory as history, match } from "react-router";
import { rootRoute } from "sourcegraph/app/App";
import * as context from "sourcegraph/app/context";
import { hashLinkScroll, shouldUpdateScroll } from "sourcegraph/app/routerScrollBehavior";
import "sourcegraph/util/actionLogger";
import { EventLogger } from "sourcegraph/util/EventLogger";

// mark files that contain only types as being used (for UnusedFilesWebpackPlugin)
import "sourcegraph/app/routeParams";
import "sourcegraph/Location";
import "sourcegraph/user";

// REQUIRED. Configures Sentry error monitoring.
import "sourcegraph/init/Sentry";

// REQUIRED. Enables HTML history API (pushState) tracking in Google Analytics.
// See https://github.com/googleanalytics/autotrack#shouldtrackurlchange.
import "autotrack/lib/plugins/url-change-tracker";

EventLogger.init();

Relay.injectNetworkLayer(new Relay.DefaultNetworkLayer("/.api/graphql", { headers: context.context.xhrHeaders }));

declare var __webpack_public_path__: any;
__webpack_public_path__ = document.head.dataset["webpackPublicPath"]; // tslint-disable-line no-undef

const rootEl = document.getElementById("main") as HTMLElement;

let hotReloadCounter = 0;

// matchWithRedirectHandling calls the router match func. If the router issues
// a redirect, it calls match recursively after replacing the location with the
// new one.
function matchWithRedirectHandling(recursed: boolean): void {
	match({ history, routes: rootRoute }, (err, redirectLocation, renderProps) => {
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
