// @flow

import {setGlobalFeatures} from "sourcegraph/app/features";
import type {Features} from "sourcegraph/app/features";

let context: {
	appURL?: string;
	authorization?: string;
	cacheControl?: string;
	currentUser?: Object;
	userEmail?: string;
	currentSpanID?: string;
	userAgentIsBot?: boolean;
	assetsRoot?: string;
	buildVars?: {Version: string};
	hasLinkedGitHub?: boolean;
} = {};

// ContextInput is the input context to set up the JS environment (e.g., from Go).
type ContextInput = typeof context & {
	// We are migrating from a global context object to using React context
	// as much as possible. These fields are only available using context wrappers.
	features?: Features;
};

// Sets the values of the context given a JSContext object from the server.
export function reset(ctx: ContextInput) {
	const features = ctx.features;
	delete ctx.features;
	if (typeof features !== "undefined") {
		setGlobalFeatures(features);
	}

	context = ctx;
}

export default context;
