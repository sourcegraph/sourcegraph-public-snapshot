// @flow weak

import React from "react";
import Dispatcher from "sourcegraph/Dispatcher";
import context from "sourcegraph/app/context";
import type {SiteConfig} from "sourcegraph/app/siteConfig";
import type {AuthInfo, User} from "sourcegraph/user";
import {getViewName, getRoutePattern} from "sourcegraph/app/routePatterns";
import type {Route} from "react-router";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import * as UserActions from "sourcegraph/user/UserActions";
import * as DefActions from "sourcegraph/def/DefActions";
import UserStore from "sourcegraph/user/UserStore";

export const EventLocation = {
	Login: "Login",
	Signup: "Signup",
	Dashboard: "Dashboard",
	DefPopup: "DefPopup",
};

export class EventLogger {
	_amplitude: any = null;
	_intercom: any = null;
	_fullStory: any = null;

	_intercomSettings: any;
	userAgentIsBot: bool;
	_dispatcherToken: any;
	_siteConfig: ?SiteConfig;

	constructor() {
		this._intercomSettings = null;

		// Listen to the UserStore for changes in the user login/logout state.
		UserStore.addListener(() => this._updateUser());

		// Listen for all Stores dispatches.
		// You must separately log "frontend" actions of interest,
		// with the relevant event properties.
		this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	setSiteConfig(siteConfig: SiteConfig) {
		this._siteConfig = siteConfig;
	}

	// init initializes Amplitude and Intercom.
	init() {
		if (global.window && !this._amplitude) {
			this._amplitude = require("amplitude-js");

			if (!this._siteConfig) {
				throw new Error("EventLogger requires SiteConfig to be previously set using EventLogger.setSiteConfig before EventLogger can be initialized.");
			}

			let apiKey = "608f75cce80d583063837b8f5b18be54";
			if (this._siteConfig.buildVars.Version === "dev") {
				apiKey = "2b4b1117d1faf3960c81899a4422a222";
			} else {
				switch (this._siteConfig.appURL) {
				case "https://sourcegraph.com":
					apiKey = "e3c885c30d2c0c8bf33b1497b17806ba";
					break;
				case "https://staging.sourcegraph.com":
				case "https://staging2.sourcegraph.com":
				case "https://staging3.sourcegraph.com":
				case "https://staging4.sourcegraph.com":
					apiKey = "903f9390c3eefd5651853cf8dbd9d363";
					break;
				default:
					break;
				}
			}

			this._amplitude.init(apiKey, null, {
				includeReferrer: true,
			});
		}

		if (global.Intercom) this._intercom = global.Intercom;
		if (global.FS) this._fullStory = global.FS;

		if (typeof window !== "undefined") {
			this._intercomSettings = window.intercomSettings;
		}

		this.userAgentIsBot = Boolean(context.userAgentIsBot);

		// Opt out of Amplitude events if the user agent is a bot.
		this._amplitude.setOptOut(this.userAgentIsBot);
	}

	// User data from the previous call to _updateUser.
	_user: ?User;
	_authInfo: AuthInfo = {};
	_primaryEmail: ?string;

	// _updateUser is be called whenever the user changes (after login or logout,
	// or on the initial page load);
	//
	// If any events have been buffered, it will flush them immediately.
	// If you do not call _updateUser or it is run on the server,
	// any subequent calls to logEvent or setUserProperty will be buffered.
	_updateUser() {
		const user = UserStore.activeUser();
		const authInfo = UserStore.activeAuthInfo();
		const emails = user && user.UID ? UserStore.emails.get(user.UID) : null;
		const primaryEmail = emails && !emails.Error ? emails.filter(e => e.Primary).map(e => e.Email)[0] : null;

		if (this._authInfo !== authInfo) {
			if (this._authInfo && this._authInfo.UID && (!authInfo || this._authInfo.UID !== authInfo.UID)) {
				// The user logged out or another user logged in on the same browser.

				// Distinguish between 2 users who log in from the same browser; see
				// https://github.com/amplitude/Amplitude-Javascript#logging-out-and-anonymous-users.
				if (this._amplitude) this._amplitude.regenerateDeviceId();

				// Prevent the next user who logs in (e.g., on a public terminal) from
				// seeing the previous user's Intercom messages.
				if (this._intercom) this._intercom("shutdown");

				if (this._fullStory) this._fullStory.clearUserCookie();
			}

			if (authInfo) {
				if (this._amplitude && authInfo.Login) this._amplitude.setUserId(authInfo.Login || null);
				if (authInfo.UID) this.setIntercomProperty("user_id", authInfo.UID.toString());
				if (authInfo.IntercomHash) this.setIntercomProperty("user_hash", authInfo.IntercomHash);
				if (this._fullStory && authInfo.Login) {
					this._fullStory.identify(authInfo.Login);
				}
			}
			if (this._intercom) this._intercom("boot", this._intercomSettings);
		}
		if (this._user !== user && user) {
			if (user.Name) this.setIntercomProperty("name", user.Name);
			if (this._fullStory) this._fullStory.setUserVars({displayName: user.Name});
			if (user.RegisteredAt) {
				this.setUserProperty("registered_at", new Date(user.RegisteredAt).toDateString());
				this.setIntercomProperty("created_at", new Date(user.RegisteredAt).getTime() / 1000);
			}
		}
		if (this._primaryEmail !== primaryEmail) {
			if (primaryEmail) {
				this.setUserProperty("email", primaryEmail);
				this.setIntercomProperty("email", primaryEmail);
				if (this._fullStory) this._fullStory.setUserVars({email: primaryEmail});
			}
		}

		this._user = user;
		this._authInfo = authInfo;
		this._primaryEmail = primaryEmail;
	}

	// sets current user's properties
	setUserProperty(property, value) {
		this._amplitude.identify(new this._amplitude.Identify().set(property, value));
	}

	// records events for the current user, if user agent is not bot
	logEvent(eventName, eventProperties) {
		if (typeof window !== "undefined" && window.localStorage["event-log"]) {
			console.debug("%cEVENT %s", "color: #aaa", eventName, eventProperties);
		}
		this._amplitude.logEvent(eventName, eventProperties);
	}

	logEventForPage(eventName, pageName, eventProperties) {
		if (!pageName) throw new Error("PageName must be defined");

		let props = eventProperties ? eventProperties : {};
		props["page_name"] = pageName;
		this.logEvent(eventName, props);
	}

	// sets current user's property value
	setIntercomProperty(property, value) {
		if (this._intercom) this._intercomSettings[property] = value;
	}

	// records intercom events for the current user
	logIntercomEvent(eventName, eventProperties) {
		if (this._intercom && !this.userAgentIsBot) this._intercom("trackEvent", eventName, eventProperties);
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.RemoteReposFetched:
			if (action.data.RemoteRepos) {
				let orgs = {};
				for (let repo of action.data.RemoteRepos) {
					if (repo.OwnerIsOrg) orgs[repo.Owner] = true;
				}
				this.setUserProperty("orgs", Object.keys(orgs));
				this.setUserProperty("num_github_repos", action.data.RemoteRepos.length);
				this.setIntercomProperty("companies", Object.keys(orgs).map(org => ({id: `github_${org}`, name: org})));
				if (orgs["sourcegraph"]) {
					this.setUserProperty("is_sg_employee", "true");
				}
			}
			break;

		case UserActions.SignupCompleted:
		case UserActions.LoginCompleted:
		case UserActions.LogoutCompleted:
		case UserActions.ForgotPasswordCompleted:
		case UserActions.ResetPasswordCompleted:
			if (action.email) {
				this.setUserProperty("email", action.email);
			}

			if (action.eventName) {
				if (action.signupChannel) {
					this.setUserProperty("signup_channel", action.signupChannel);
					this.logEvent(action.eventName, {error: Boolean(action.resp.Error), signup_channel: action.signupChannel});
				} else {
					this.logEvent(action.eventName, {error: Boolean(action.resp.Error)});
				}
			}
			break;

		case DefActions.DefsFetched:
			if (action.eventName) {
				let eventProps = {
					query: action.query,
					overlay: action.overlay,
				};
				this.logEvent(action.eventName, eventProps);
			}
			break;
		default:
			// All dispatched actions to stores will automatically be tracked by the eventName
			// of the action (if set). Override this behavior by including another case above.
			if (action.eventName) {
				this.logEvent(action.eventName);
			}
			break;
		}
	}
}

export default new EventLogger();

// withEventLoggerContext makes eventLogger accessible as this.context.eventLogger
// in the component's context.
export function withEventLoggerContext(eventLogger: EventLogger, Component: ReactClass): ReactClass {
	class WithEventLogger extends React.Component {
		static childContextTypes = {
			eventLogger: React.PropTypes.object,
		};

		constructor(props) {
			super(props);
			eventLogger.init();
		}

		getChildContext(): {eventLogger: EventLogger} {
			return {eventLogger};
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithEventLogger;
}

// withViewEventsLogged calls this.context.eventLogger.logEvent when the
// location's pathname changes.
export function withViewEventsLogged(Component: ReactClass): ReactClass {
	class WithViewEventsLogged extends React.Component { // eslint-disable-line react/no-multi-comp
		static propTypes = {
			routes: React.PropTypes.arrayOf(React.PropTypes.object),
			location: React.PropTypes.object.isRequired,
		};

		static contextTypes = {
			router: React.PropTypes.object.isRequired,
			eventLogger: React.PropTypes.object.isRequired,
		};

		componentDidMount() {
			this._logView(this.props.routes, this.props.location);
			this._checkEventQuery();
		}

		componentWillReceiveProps(nextProps) {
			// Greedily log page views. Technically changing the pathname
			// may match the same "view" (e.g. interacting with the directory
			// tree navigations will change your URL,  but not feel like separate
			// page events). We will log any change in pathname as a separate event.
			// NOTE: this will not log separate page views when query string / hash
			// values are updated.
			if (this.props.location.pathname !== nextProps.location.pathname) {
				this._logView(nextProps.routes, nextProps.location);
			}

			this._checkEventQuery();
		}

		camelCaseToUnderscore(input) {
			if (input.charAt(0) === "_") {
				input = input.substring(1);
			}

			return input.replace(/([A-Z])/g, function($1) {
				return `_${$1.toLowerCase()}`;
			});
		}

		_checkEventQuery() {
			// Allow tracking events that occurred externally and resulted in a redirect
			// back to Sourcegraph. Pull the event name out of the URL.
			if (this.props.location.query && this.props.location.query._event) {
				// For login signup related metrics a channel will be associated with the signup.
				// This ensures we can track one metrics "SignupCompleted" and then query on the channel
				// for more granular metrics.
				let eventProperties= {};
				for (let key in this.props.location.query) {
					if (key !== "_event") {
						eventProperties[this.camelCaseToUnderscore(key)] = this.props.location.query[key];
					}
				}

				if (this.props.location.query._githubAuthed) {
					this.context.eventLogger.setUserProperty("github_authed", this.props.location.query._githubAuthed);
				}

				this.context.eventLogger.logEvent(this.props.location.query._event, eventProperties);

				// Won't take effect until we call replace below, but prevents this
				// from being called 2x before the setTimeout block runs.
				delete this.props.location.query._event;
				delete this.props.location.query._githubAuthed;

				// Remove _event from the URL to canonicalize the URL and make it
				// less ugly.
				const locWithoutEvent = {...this.props.location,
					query: {...this.props.location.query, _event: undefined, _signupChannel: undefined, _onboarding: undefined, _githubAuthed: undefined}, // eslint-disable-line no-undefined
					state: {...this.props.location.state, _onboarding: this.props.location.query._onboarding},
				};

				delete this.props.location.query._signupChannel;
				delete this.props.location.query._onboarding;

				this.context.router.replace(locWithoutEvent);
			}
		}

		_logView(routes: Array<Route>, location: Location) {
			let eventProps = {
				url: location.pathname,
			};
			// TODO:matt remove this once all plugins are switched to new version
			// This is temporarily here for backwards compat
			if (location.query && location.query["utm_source"] === "chromeext") {
				eventProps = {
					referred_by_browser_ext: "chrome",
					url: location.pathname,
				};
			} else if (location.query && location.query["utm_source"] === "browser-ext" && location.query["browser_type"]) {
				eventProps = {
					referred_by_browser_ext: location.query["browser_type"],
					url: location.pathname,
				};
			} else if (location.query && location.query["utm_source"] === "sourcegraph-editor" && location.query["editor_type"]) {
				eventProps = {
					url: location.pathname,
					referred_by_sourcegraph_editor: location.query["editor_type"],
				};
			}

			const viewName = getViewName(routes);
			if (viewName) {
				this.context.eventLogger.logEvent(viewName, eventProps);
			} else {
				this.context.eventLogger.logEvent("UnmatchedRoute", {
					...eventProps,
					pattern: getRoutePattern(routes),
				});
			}
		}

		render() { return <Component {...this.props} />; }
	}
	return WithViewEventsLogged;
}
