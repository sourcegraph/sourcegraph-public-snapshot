const context = {
	appURL: "",
	authorization: "",
	currentUser: null,
	csrfToken: "",
	cacheControl: null,
	parentSpanID: null,
	assetsRoot: "",
	buildVars: null,
	userAgent: "",
};

// Sets the values of the context given a JSContext object from the server.
//
// TODO(pure-react) Type check this.
export function reset(ctx) {
	context.appURL = ctx.AppURL;
	context.authorization = ctx.Authorization;
	context.currentUser = ctx.CurrentUser;
	context.csrfToken = ctx.CSRFToken;
	context.cacheControl = ctx.CacheControl || null;
	context.currentSpan = ctx.CurrentSpanID;
	context.assetsRoot = ctx.AssetsRoot;
	context.buildVars = ctx.BuildVars;
	context.userAgent = ctx.UserAgent;
}

export default context;
