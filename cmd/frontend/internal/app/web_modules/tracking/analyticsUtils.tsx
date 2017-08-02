import * as URI from "urijs";
import { eventLogger } from "app/tracking/eventLogger";
import { events } from "app/tracking/events";
import { EventCategories, EventActions } from "app/tracking/analyticsConstants";

/**
 * Get pageview-specific event properties from URL query string parameters
 */
export function pageViewQueryParameters(url: string) {
	const parsedUrl = URI.parse(url);
	const query = URI.parseQuery(parsedUrl.query);
	return {
		utm_campaign: query['utm_campaign'],
		utm_source: query['utm_source'],
		utm_product_name: query['utm_product_name'],
		utm_product_version: query['utm_product_version']
	};
}

/**
 * Log events associated with URL query string parameters, and remove those parameters as necessary
 * Note that this is a destructive operation (it changes the page URL and replaces browser state) by
 * calling stripURLParameters
 */
export function handleQueryEvents(url: string) {
	const parsedUrl = URI.parse(url);
	const query = URI.parseQuery(parsedUrl.query);
	const eventParameters = Object.keys(query)
		.reduce((r, key) => {
			r[camelCaseToUnderscore(key)] = query[key];
			return r;
		}, {});
	const eventName = query['_event'];
	const isBadgeRedirect = query['badge'] !== undefined;

	// TODO(Dan): add handling for new auth scheme
	if (eventName || isBadgeRedirect) {
		if (isBadgeRedirect) {
			events.RepoBadgeRedirected.log(eventParameters);
		} else if (eventName) {
			eventLogger.logEvent(EventCategories.External, EventActions.Redirect, eventName, eventParameters);
		}
	}

	stripURLParameters(url, [
		'_event',
		'_source',
		'utm_campaign',
		'utm_source',
		'utm_product_name',
		'utm_product_version',
		'badge'
	]);
}

/**
 * Strip provided URL parameters and update window history
 */
function stripURLParameters(url: string, paramsToRemove: string[] = []): void {
	const parsedUrl = URI.parse(url);
	const currentQuery = URI.parseQuery(parsedUrl.query);
	const newQuery = Object.keys(currentQuery)
		.filter(key => { return paramsToRemove.indexOf(key) === -1; })
		.reduce((r, key) => {
			r[key] = currentQuery[key];
			return r;
		}, {});
	parsedUrl.query = URI.buildQuery(newQuery);
	window.history.replaceState({}, window.document.title, URI.build(parsedUrl));
}

function camelCaseToUnderscore(input: string): string {
	if (input.charAt(0) === '_') {
		input = input.substring(1);
	}
	return input.replace(/([A-Z])/g, ($1) => `_${$1.toLowerCase()}`);
}
