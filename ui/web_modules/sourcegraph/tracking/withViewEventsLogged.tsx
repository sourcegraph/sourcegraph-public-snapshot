import * as React from "react";
import { Route } from "react-router";
import { getRoutePattern, getViewName } from "sourcegraph/app/routePatterns";
import { Router, RouterLocation } from "sourcegraph/app/router";
import { Events, LogUnknownRedirectEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";
import { getLanguageExtensionForPath } from "sourcegraph/util/inventory";

// withViewEventsLogged calls (this.context as any).eventLogger.logEvent when the
// location's pathname changes.
interface WithViewEventsLoggedProps {
	routes: Route[];
	location: RouterLocation;
}

export function withViewEventsLogged<P extends WithViewEventsLoggedProps>(component: React.ComponentClass<{}>): React.ComponentClass<{}> {
	class WithViewEventsLogged extends React.Component<P, {}> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object.isRequired,
		};

		context: {
			router: Router,
		};

		componentDidMount(): void {
			this.logViewAsync(this.props.routes, this.props.location);
			this._checkEventQuery();
		}

		componentWillReceiveProps(nextProps: P): void {
			// Greedily log page views. Technically changing the pathname
			// may match the same "view" (e.g. interacting with the directory
			// tree navigations will change your URL,  but not feel like separate
			// page events). We will log any change in pathname as a separate event.
			// NOTE: this will not log separate page views when query string / hash
			// values are updated.
			if (this.props.location.pathname !== nextProps.location.pathname) {
				this.logViewAsync(nextProps.routes, nextProps.location);
				// Greedily update the event logging tracker identity
				EventLogger.updateTrackerWithIdentificationProps();
			}

			this._checkEventQuery();
		}

		camelCaseToUnderscore(input: string): string {
			if (input.charAt(0) === "_") {
				input = input.substring(1);
			}

			return input.replace(/([A-Z])/g, ($1) => `_${$1.toLowerCase()}`);
		}

		_checkEventQuery(): void {
			// Allow tracking events that occurred externally and resulted in a redirect
			// back to Sourcegraph. Pull the event name out of the URL.
			const eventName = this.props.location.query["_event"];
			const isBadgeRedirect = this.props.location.query["badge"] !== undefined;

			if (this.props.location.query && (eventName || isBadgeRedirect)) {
				// For login signup related metrics a channel will be associated with the signup.
				// This ensures we can track one metrics "SignupCompleted" and then query on the channel
				// for more granular metrics.
				let eventProperties = {};
				for (let key in this.props.location.query) {
					if (key !== "_event" && key !== "badge") {
						eventProperties[this.camelCaseToUnderscore(key)] = this.props.location.query[key];
					}
				}

				if (this.props.location.query["_githubAuthed"]) {
					EventLogger.setUserIsGitHubAuthed(this.props.location.query["_githubAuthed"]);
					if (eventName === Events.Signup_Completed.label) {
						Events.Signup_Completed.logEvent(eventProperties);
					} else if (eventName === Events.OAuth2FlowGitHub_Completed.label) {
						Events.OAuth2FlowGitHub_Completed.logEvent(eventProperties);
					}
				} else if (this.props.location.query["_invited_by_user"] || this.props.location.query["_org_invite"]) {
					EventLogger.setUserInvited(this.props.location.query["_invited_by_user"] || "", this.props.location.query["_org_invite"] || "");
					Events.OrgEmailInvite_Clicked.logEvent(eventProperties);
				} else if (eventName === Events.RepoBadge_Redirected.label) {
					Events.RepoBadge_Redirected.logEvent(eventProperties);
				} else if (eventName) {
					LogUnknownRedirectEvent(eventName, eventProperties);
				} else if (isBadgeRedirect) {
					Events.RepoBadge_Redirected.logEvent(eventProperties);
				}

				// Won't take effect until we call replace below, but prevents this
				// from being called 2x before the setTimeout block runs.
				delete this.props.location.query["_event"];
				delete this.props.location.query["_githubAuthed"];
				delete this.props.location.query["_org_invite"];
				delete this.props.location.query["_invited_by_user"];
				delete this.props.location.query["_def_info_def"];
				delete this.props.location.query["_repo"];
				delete this.props.location.query["_rev"];
				delete this.props.location.query["_path"];
				delete this.props.location.query["_source"];
				delete this.props.location.query["_githubCompany"];
				delete this.props.location.query["_githubName"];
				delete this.props.location.query["_githubLocation"];
				delete this.props.location.query["_signupChannel"];
				delete this.props.location.query["_onboarding"];
				delete this.props.location.query["badge"];

				// Remove _event from the URL to canonicalize the URL and make it
				// less ugly.
				const locWithoutEvent = Object.assign({}, this.props.location, {
					query: Object.assign({}, this.props.location.query, { _event: undefined, _signupChannel: undefined, _onboarding: undefined, _githubAuthed: undefined, _invited_by_user: undefined, _org_invite: undefined, _def_info_def: undefined, _repo: undefined, _rev: undefined, _path: undefined, _source: undefined, _githubCompany: undefined, _githubName: undefined, _githubLocation: undefined, badge: undefined }),
					state: Object.assign({}, this.props.location.state, { _onboarding: this.props.location.query["_onboarding"] }),
				});

				(this.context as any).router.replace(locWithoutEvent);
			}
		}

		private logViewAsync(routes: Route[], location: RouterLocation): void {
			setTimeout(() => {
				// Logging to Telligent takes a long time and blocks us from
				// rendering the component, so do it async.
				this.logView(routes, location);
			});
		}

		private logView(routes: Route[], location: RouterLocation): void {
			let eventProps: {
				url: string;
				language?: string;
				utm_campaign?: string;
				utm_source?: string;
			};

			if (location.query && (location.query["utm_campaign"] || location.query["utm_source"])) {
				eventProps = Object.assign({}, { url: location.pathname },
					location.query["utm_campaign"] ? { utm_campaign: location.query["utm_campaign"] } : {},
					location.query["utm_source"] ? { utm_source: location.query["utm_source"] } : {});
			} else {
				eventProps = {
					url: location.pathname,
				};
			}

			const viewName = getViewName(routes);
			const routeParams = this.context.router.params;

			if (viewName) {
				if (viewName === "ViewBlob" && routeParams) {
					const filePath = routeParams.splat[routeParams.splat.length - 1];
					const lang = getLanguageExtensionForPath(filePath);
					if (lang) { eventProps.language = lang; }
				}
				EventLogger.logViewEvent(viewName, location.pathname, Object.assign({}, eventProps, { pattern: getRoutePattern(routes) }));
			} else {
				EventLogger.logViewEvent("UnmatchedRoute", location.pathname, Object.assign({}, eventProps, { pattern: getRoutePattern(routes) }));
			}
		}

		render(): JSX.Element | null {
			return React.createElement(component, this.props);
		}
	}
	return WithViewEventsLogged;
}
