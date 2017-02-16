import { EventLogger } from "../tracking/EventLogger";
import { ExtensionEventLogger } from "../tracking/ExtensionEventLogger";

let eventLogger = new ExtensionEventLogger();

export function getEventLogger(): EventLogger | undefined {
	return eventLogger;
}

export function getPlatformName(): string {
	return window.navigator.userAgent.indexOf("Firefox") !== -1 ? "firefox-extension" : "chrome-extension";
}

export function getDomain(loc: Location): Domain {
	if (Boolean(/^https?:\/\/phabricator.aws.sgdev.org/.test(loc.href))) {
		return Domain.SGDEV_PHABRICATOR;
	}
	if (Boolean(/^https?:\/\/(www.)?github.com/.test(loc.href))) {
		return Domain.GITHUB;
	}
	if (Boolean(/^https?:\/\/(www.)?sourcegraph.com/.test(loc.href))) {
		return Domain.SOURCEGRAPH;
	}
	throw new Error(`Unable to determine the domain, ${loc.href}`);
}

export enum Domain {
	GITHUB,
	SGDEV_PHABRICATOR,
	SOURCEGRAPH,
}
