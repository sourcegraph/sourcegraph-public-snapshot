const DEFAULT_SENTRY_OPTS = {
	level: "error",
};

export function logToSentry(errorMessage: string, logToConsole: boolean = true, sentryOpts?: {}): void {
	const sentryOptions = { ...DEFAULT_SENTRY_OPTS, ...sentryOpts };
	// only log to Sentry if this is a production environment
	if (isProduction() && global && global.window && global.window.Raven) {
		global.window.Raven.captureMessage(errorMessage, sentryOptions);
	}
	// however, even if non production we can log to console
	if (logToConsole) {
		console.error(errorMessage);
	}
}

function isProduction(): boolean {
	return typeof window !== undefined && process.env["NODE_ENV"] === "production";
}
