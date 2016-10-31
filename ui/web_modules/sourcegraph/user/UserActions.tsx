import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export type Action =
	SubmitBetaSubscription |
	BetaSubscriptionCompleted

export class SubmitBetaSubscription {
	email: string;
	firstName: string;
	lastName: string;
	languages: string[];
	editors: string[];
	message: string;

	constructor(email: string, firstName: string, lastName: string, languages: string[], editors: string[], message: string) {
		this.email = email;
		this.firstName = firstName;
		this.lastName = lastName;
		this.languages = languages;
		this.editors = editors;
		this.message = message;
	}
}

export class BetaSubscriptionCompleted {
	resp: any;
	eventObject: AnalyticsConstants.LoggableEvent;

	constructor(resp: any) {
		this.resp = resp;
		this.eventObject = AnalyticsConstants.Events.BetaSubscription_Completed;
	}
}
