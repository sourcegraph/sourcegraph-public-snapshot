// tslint:disable-next-line
/// <reference path="../../../node_modules/@types/google.analytics/index.d.ts" />

import { Features } from "sourcegraph/util/features";

class GoogleAnalyticsWrapper {
	ga: UniversalAnalytics.ga | null;
	gaClientID: string | null;
	constructor() {
		if (global && global.window && global.window.ga) {
			this.ga = global.window.ga;
		} else {
			return;
		}
		// required because TypeScript says so
		if (!this.ga) {
			return;
		}
		this.ga(this.setTrackerID);
	}

	private setTrackerID = (tracker: any) => {
		if (Features.eventLogDebug.isEnabled()) {
			// TODO(uforic): remove after bug is resolved
			/* tslint:disable */
			console.log("Setting google analytics tracking id");
			console.log(tracker.get("clientId"));
			console.log(tracker);
			/* tslint:enable */
		}
		this.gaClientID = tracker.get("clientId");
	}

	setTrackerLogin(loginInfo: string): void {
		if (!this.ga) {
			return;
		}
		this.ga("set", "userId", loginInfo);
	}

	logEventCategoryComponents(eventCategory: string, eventAction: string, eventLabel: string, nonInteraction: boolean = false): void {
		if (!this.ga) {
			return;
		}
		// not sure that eventCatrgory or eventAction can be null here.
		this.ga("send", {
			hitType: "event",
			eventCategory: eventCategory || "",
			eventAction: eventAction || "",
			eventLabel: eventLabel,
			nonInteraction: nonInteraction,
		});
	}

}

export const googleAnalytics = new GoogleAnalyticsWrapper();
