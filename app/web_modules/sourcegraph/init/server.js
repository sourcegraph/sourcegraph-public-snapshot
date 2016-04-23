import "babel-polyfill";
import React from "react";
import Helmet from "react-helmet";
import ReactDOMServer from "react-dom/server";
import {RouterContext, match, createMemoryHistory} from "react-router";
import {rootRoute} from "sourcegraph/app/App";
import dumpStores from "sourcegraph/init/dumpStores";
import resetStores from "sourcegraph/init/resetStores";
import {trackedPromisesCount, allTrackedPromisesResolved} from "sourcegraph/app/status";
import * as context from "sourcegraph/app/context";
import split from "split";
import {httpStatusCode} from "sourcegraph/app/status";

/*
TODO: We can optimize this iterative rendering scheme to avoid wasted effort
in actual rendering when we know we don't have sufficient data to be finished.
We need to do something like throwing an error when any fetch is called, and
catching that error in the call below to renderToString.

This would mean that we'd have more render iterations (since each new iteration
would only trigger at most 1 additional async fetch), but they would be cheaper
since they would not do any actual rendering, only component initialization.
*/

// renderIter iteratively renders the HTML so that fetches are triggered. Do this
// until we reach the fixed point where no additional fetches are triggered.
function renderIter(i, props, location, deadline, callback) {
	if (i > 10) {
		throw new Error(`Maximum React server-side rendering iterations reached (${i}).`);
	}

	let t0 = Date.now();

	// Trigger a render so that we start the async data fetches.
	let htmlStr;
	try {
		htmlStr = ReactDOMServer.renderToString(<RouterContext {...props} />);
		if (global.process.env.DEBUGUI) console.log(`RENDER#${i}: renderToString took ${Date.now() - t0} msec`);
	} catch (e) {
		return Promise.reject(e);
	}

	if (trackedPromisesCount() === 0) {
		if (i > 1) {
			if (global.process.env.DEBUGUI) console.warn(`PERF NOTE: Rendering path ${props.location.pathname} took ${i} iterations (of renderToString and server RTTs) due to new async fetches being triggered after each iteration (likely as more data became available). Pipeline data better to improve performance.`);
		}

		// No additional async fetches were triggered, so we are done!
		const head = Helmet.rewind();
		return Promise.resolve({
			statusCode: location.state ? httpStatusCode(location.state.error) : 200,
			body: htmlStr,
			contentType: "text/html; charset=utf-8",
			stores: dumpStores(),
			cache: location.state && location.state.cache,
			head: {
				htmlAttributes: head.htmlAttributes.toString(),
				title: head.title.toString(),
				base: head.base.toString(),
				meta: head.meta.toString(),
				link: head.link.toString(),
				script: head.script.toString(),
			},
		});
	}

	let tFetch0 = Date.now();
	const promisesCount = trackedPromisesCount();
	return allTrackedPromisesResolved().then(() => {
		if (global.process.env.DEBUGUI) console.log(`RENDER#${i}: ${promisesCount} fetches took ${Date.now() - tFetch0} msec`);
		return renderIter(i + 1, props, location, deadline, callback);
	});
}

function resetAll(arg) {
	// SECURITY NOTE: You must clear out any state so that each time handle is
	// called, no data or credentials from the previous request remain. Otherwise
	// handle may return data that the current user is unauthorized to see (but
	// that was processed during a previous request for a different user).
	//
	// It's also CRUCIAL that we exit the process (see the process.exit call below)
	// if there is a failure, since there might be other in-flight async operations
	// that could alter global state after the failure is reported.
	context.reset(arg.jsContext);
	resetStores();
}

// handle is called from Go to render the page's contents.
const handle = (arg, callback) => {
	resetAll(arg);

	// Track the current location.
	let hist = createMemoryHistory(arg.location);
	let location = {};
	hist.listen((loc) => {
		Object.assign(location, loc);
	});

	match({history: hist, location: arg.location, routes: rootRoute}, (err, redirectLocation, renderProps) => {
		if (typeof err === "undefined" && typeof redirectLocation === "undefined" && typeof renderProps === "undefined") {
			callback({
				statusCode: 404,
				contentType: "text/plain",
				error: "no route",
			});
			return;
		}
		if (err) {
			callback({
				statusCode: 500,
				error: `Routing error: ${err}`,
			});
			return;
		}
		if (redirectLocation) {
			// Assumes that all redirects are 301s (Moved Permanently).
			callback({
				statusCode: 301,
				redirectLocation: redirectLocation.href,
				contentType: "text/html",
				body: "Redirecting...",
			});
			return;
		}

		const props = {...renderProps, ...arg.extraProps};

		renderIter(1, props, location, arg.deadline, callback)
			.catch((e) => ({
				statusCode: 500,
				error: `Uncaught server-side rendering error:\n${e.stack}`,
			}))
			.then((resp) => callback(resp));
	});
};

// jsserver: listens on stdin for lines of JSON sent by the app/internal/ui Go package.
if (typeof global !== "undefined" && global.process && global.process.env.JSSERVER) {
	global.process.stdout.write("\"ready\"\n");
	console.log = console.error;

	process.stdin.pipe(split())
		.on("data", (line) => {
			if (line === "") return;
			handle(JSON.parse(line), (data) => {
				global.process.stdout.write(JSON.stringify(data));
				global.process.stdout.write("\n");
				if (data.error) {
					// Exit so that we don't reuse this process and its potentially
					// corrupted state.
					global.process.stderr.write("jsserver process exiting due to error:\n");
					global.process.stderr.write(data.error);
					global.process.exit(1);
				}
			});
		})
		.on("error", (err) => {
			console.error("jsserver: error reading line from stdin:", err);
		});
}
