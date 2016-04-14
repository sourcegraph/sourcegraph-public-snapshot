// @flow

const context: {
	appURL?: string;
	authorization?: string;
	cacheControl?: string;
	currentUser?: Object;
	userEmail?: string;
	currentSpanID?: string;
	userAgent?: string;
	assetsRoot?: string;
	buildVars?: Object;
	features?: {[key: string]: any};
} = {};

// Sets the values of the context given a JSContext object from the server.
export function reset(ctx: typeof context) {
	Object.assign(context, ctx);
}

export default context;
