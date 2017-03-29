import { Events, LoggableEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { checkStatus, defaultFetch as fetch } from "sourcegraph/util/xhr";

interface RawFormPayload {
	[key: string]: string | string[] | undefined;
}
interface FormPayload {
	[key: string]: string | undefined;
}

/** 
 * submitHubSpotForm prepares and submits data to HubSpot
 */
function submitHubSpotForm(hubSpotFormName: string, body: RawFormPayload, successEvent?: LoggableEvent): Promise<Response> {
	const payload = prepareFormPayload(body);
	return fetch(`/.api/submit-form`, {
		method: "POST",
		headers: { "Content-Type": "application/json; charset=utf-8" },
		body: JSON.stringify({ ...payload, hubSpotFormName: hubSpotFormName }),
	})
		.then(checkStatus)
		// Fetch the response's body as an object using the json() method to return a Promise
		.then(resp => resp.json())
		.then((resp: any) => {
			if (successEvent !== undefined) {
				successEvent.logEvent({ form: { ...payload, hubSpotFormName: hubSpotFormName, response: resp } });
			}
			return resp;
		})
		.catch(err => {
			// Wrap the server error message
			throw Object.assign(err, { message: `An error occurred while submitting this form: ${err.message}` });
		});
}

/** 
 * Beta form submissions
 */
interface BetaSignupPayload extends RawFormPayload {
	beta_email: string;
	firstname: string;
	lastname: string;
	company: string;
	languages_used: string[];
	editors_used: string[];
	message: string;
}
export function submitBetaSignupForm(body: BetaSignupPayload): Promise<Response> {
	return submitHubSpotForm("BetaSignupForm", body, Events.BetaSubscription_Completed);
}

/** 
 * Zap beta form submissions
 */
interface ZapBetaSignupPayload extends RawFormPayload {
	beta_email: string;
	firstname: string;
	lastname: string;
	company: string;
	editors_used: string[];
}
export function submitZapBetaSignupForm(body: ZapBetaSignupPayload): Promise<Response> {
	return submitHubSpotForm("ZapBetaSignupForm", body, Events.ZapBetaSubscription_Completed);
}

/** 
 * Change user plan form submissions
 */
interface ChangeUserPlanPayload extends RawFormPayload {
	changePlanMessage: string;
}
export function submitChangeUserPlanForm(body: ChangeUserPlanPayload): Promise<Response> {
	return submitHubSpotForm("ChangeUserPlan", body);
}

/** 
 * After-signup form submissions 
 */
interface AfterSignupPayload extends RawFormPayload {
	firstname: string;
	lastname: string;
	company?: string;
	signupEmail: string;
	plan: string;
	isPrivateCodeUser: string;
	githubOrgs?: string;
	existingSoftware?: string;
	versionControlSystem?: string;
	numberOfDevs?: string;
	otherDetails?: string;
}
export function submitAfterSignupForm(body: AfterSignupPayload): Promise<Response> {
	return submitHubSpotForm("AfterSignupForm", body);
}

/**
 * prepareFormPayload converst a raw payload (which can contain string arrays)
 * into a HubSpot-compatible payload (strings only) in a consistent way for all forms
 */
function prepareFormPayload(payload: RawFormPayload): FormPayload {
	const outputPayload: FormPayload = {};
	for (const key of Object.keys(payload)) {
		const value = payload[key];
		if (value instanceof Array) {
			outputPayload[key] = value.join(",") + ",";
		} else if (typeof payload[key] === "string") {
			outputPayload[key] = value;
		}
	}
	return outputPayload;
}
