// @flow

import {setGlobalFeatures} from "sourcegraph/app/features";
import type {Features} from "sourcegraph/app/features";
import {setGlobalSiteConfig} from "sourcegraph/app/siteConfig";

let context: {
	authorization?: string;
	csrfToken?: string;
	cacheControl?: string;
	currentUser?: Object;
	userEmail?: string;
	currentSpanID?: string;
	userAgentIsBot?: boolean;
	hasLinkedGitHub?: boolean;
} = {};

// ContextInput is the input context to set up the JS environment (e.g., from Go).
type ContextInput = typeof context & {
	// We are migrating from a global context object to using React context
	// as much as possible. These fields are only available using context wrappers.
	appURL?: string;
	assetsRoot?: string;
	buildVars?: {Version: string};
	features?: Features;
};

// Sets the values of the context given a JSContext object from the server.
export function reset(ctx: ContextInput) {
	const features = ctx.features;
	delete ctx.features;
	if (typeof features !== "undefined") {
		setGlobalFeatures(features);
	}

	const {appURL, assetsRoot, buildVars} = ctx;
	if (typeof appURL === "undefined" || typeof assetsRoot === "undefined" || typeof buildVars === "undefined") {
		throw new Error("appURL, assetsRoot, and buildVars must all be set");
	}
	setGlobalSiteConfig({appURL, assetsRoot, buildVars});
	delete ctx.appURL;
	delete ctx.assetsRoot;
	delete ctx.buildVars;

	// $FlowHack
	Object.assign(context, ctx);
}

export default context;
