// @flow

const context: {
	appURL?: string;
	authorization?: string;
	cacheControl?: string;
	currentUser?: Object;
	currentSpanID?: Object;
	userAgent?: string;
	assetsRoot?: string;
	buildVars?: Object;
	features?: Object;
} = {};

// Sets the values of the context given a JSContext object from the server.
export function reset(ctx: typeof context) {
	Object.assign(context, ctx);
}

export default context;
