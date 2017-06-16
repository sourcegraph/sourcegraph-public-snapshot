import { EventLogger } from "../tracking/EventLogger";
import { PhabricatorInstance } from "./classes";

export let eventLogger: EventLogger;

export function setEventLogger(logger: EventLogger): void {
	if (eventLogger) {
		console.error(`event logger is being set twice, currently is ${eventLogger} and being set to ${logger}`);
	}
	eventLogger = logger;
}

// TODO(john): fix initialization (without setting this, the background script cannot scrape cookies
// because the Sourcegraph url is undefined).
export let sourcegraphUrl: string = "https://sourcegraph.com";

export let searchEnabled: boolean = false;

export function setSourcegraphUrl(url: string): void {
	sourcegraphUrl = url;
}

export let phabricatorInstance: PhabricatorInstance;

export function setPhabricatorInstance(instance: PhabricatorInstance): void {
	phabricatorInstance = instance;
}

export function isBrowserExtension(): boolean {
	return !phabricatorInstance;
}

export function setSearchEnabled(enabled: boolean): void {
	searchEnabled = enabled;
}
