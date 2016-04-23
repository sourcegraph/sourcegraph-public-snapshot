// @flow

import {setGlobalFeatures} from "sourcegraph/app/features";
import type {Features} from "sourcegraph/app/features";

const context: {
	appURL?: string;
	authorization?: string;
	cacheControl?: string;
	currentUser?: Object;
	userEmail?: string;
	currentSpanID?: string;
	userAgentIsBot?: boolean;
	assetsRoot?: string;
	buildVars?: Object;
	hasLinkedGitHub?: boolean;

	// We are migrating from a global context object to using React context
	// as much as possible. These fields are only available using context wrappers.
	features?: Features;
} = {};

// Sets the values of the context given a JSContext object from the server.
export function reset(ctx: typeof context) {
	const features = ctx.features;
	if (typeof features !== "undefined") {
		delete ctx.features;
		setGlobalFeatures(features);
	}

	Object.assign(context, ctx);
}

export default context;
