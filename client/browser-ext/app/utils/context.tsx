import { EventLogger } from "../tracking/EventLogger";
import { ExtensionEventLogger } from "../tracking/ExtensionEventLogger";

export let eventLogger = new ExtensionEventLogger();

export function getPlatformName(): string {
	return window.navigator.userAgent.indexOf("Firefox") !== -1 ? "firefox-extension" : "chrome-extension";
}

export function getSourcegraphUrl(): string {
	return "https://sourcegraph.com";
}
