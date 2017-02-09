const DEFAULT_SENTRY_OPTS = {
	level: "error",
};

interface MessageOptions {
	showInDevelopment: boolean;
	showInProduction: boolean;
};

const DEFAULT_MESSAGE_OPTS = {
	showInDevelopment: false,
	showInProduction: true,
};

export function logToSentry(errorMessage: string, messageOpts?: MessageOptions, sentryOpts?: {}): void {
	const sentryOptions = { ...DEFAULT_SENTRY_OPTS, ...sentryOpts };
	const messageOptions = { ...DEFAULT_MESSAGE_OPTS, ...messageOpts };
	if (global && global.window && global.window.Raven) {
		global.window.Raven.captureMessage(errorMessage, sentryOptions);
	}
	if (typeof window !== undefined && process.env["NODE_ENV"] === "test") {
		return;
	}
	if (process.env["NODE_ENV"] === "development" && !messageOptions.showInDevelopment) {
		return;
	}
	if (process.env["NODE_ENV"] === "production" && !messageOptions.showInProduction) {
		return;
	}
	console.error(errorMessage);
}
