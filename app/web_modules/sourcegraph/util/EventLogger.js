import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import context from "sourcegraph/app/context";

import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as UserActions from "sourcegraph/user/UserActions";

export class EventLogger {
	constructor() {
		this._amplitude = null;

		// Listen for all Stores dispatches.
		// You must separately log "frontend" actions of interest,
		// with the relevant event properties.
		this.dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	// reset() receives any event data which is buffered
	// during server-side rendering; this data will
	// be flushed to Amplitude after the first call to
	// init() in the browser.
	reset(data) {
		this.events = deepFreeze(data && data.events ? data.events : []);
		this.userProperties = deepFreeze(data && data.userProperties ? data.userProperties : []);
	}
	toJSON() {
		return {
			events: this.events,
			userProperties: this.userProperties,
		};
	}

	// Loads the Amplitude JavaScript SDK if this
	// code is run in the browser (i.e. not with node
	// when doing server-side rendering.) If any events
	// have been buffered, it will flush them immediately.
	// If you do not call init() or it is run on the server,
	// any subequent calls to logEvent or setUserProperty
	// will be buffered.
	init() {
		if (global.window && !this._amplitude) {
			this._amplitude = require("amplitude-js");

			let user = null;
			if (context.currentUser) {
				user = context.currentUser.Login;
			}

			let apiKey = "608f75cce80d583063837b8f5b18be54";
			if (context.buildVars.Version === "dev") {
				apiKey = "2b4b1117d1faf3960c81899a4422a222";
			} else {
				switch (context.appURL) {
				case "https://sourcegraph.com":
					apiKey = "e3c885c30d2c0c8bf33b1497b17806ba";
					break;
				case "https://staging.sourcegraph.com":
					apiKey = "903f9390c3eefd5651853cf8dbd9d363";
					break;
				default:
					break;
				}
			}

			this._amplitude.init(apiKey, user, {
				includeReferrer: true,
			});

			if (context.currentUser && context.currentUser.RegisteredAt) {
				this.setUserProperty("registered_at", new Date(context.currentUser.RegisteredAt).toDateString());
			}
		}

		this.isUserAgentBot = false;
		if (context.userAgent) {
			if (decodeURIComponent(context.userAgent)
				.match(/googlecloudmonitoring|pingdom.com|go .* package http|sourcegraph e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|sourcegraph-client/)) {
				this.isUserAgentBot = true;
			}
		}

		if (this._shouldFlush()) {
			this._flush();
		}
	}

	// Only flush events on the client, after a call to init().
	// Filter out bot / test user agents.
	_shouldFlush() {
		return Boolean(this._amplitude) && !this.isUserAgentBot;
	}

	// sets current context's user properties
	setUserProperty(property, value) {
		if (!this._shouldFlush()) {
			this.userProperties = deepFreeze(this.userProperties.concat([[property, value]]));
		} else {
			this._amplitude.identify(new this._amplitude.Identify().set(property, value));
		}
	}

	// records events for the current context's user
	logEvent(eventName, eventProperties) {
		if (!this._shouldFlush()) {
			this.events = deepFreeze(this.events.concat([[eventName, eventProperties]]));
		} else {
			this._amplitude.logEvent(eventName, eventProperties);
		}
	}

	_flush() {
		if (this._shouldFlush()) { // sanity check
			if (this.events) {
				for (let tuple of this.events) {
					this.logEvent(tuple[0], tuple[1]);
				}
				this.events = deepFreeze([]);
			}
			if (this.userProperties) {
				for (let tuple of this.userProperties) {
					this.setUserProperty(tuple[0], tuple[1]);
				}
				this.userProperties = deepFreeze([]);
			}
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.HomeFetched:
			if (action.data.HasLinkedGitHub && action.data.RemoteRepos) {
				let orgs = {};
				for (let repo of action.data.RemoteRepos) {
					if (repo.OwnerIsOrg) orgs[repo.Owner] = true;
				}
				this.setUserProperty("orgs", Object.keys(orgs));
				this.setUserProperty("num_github_repos", action.data.RemoteRepos.length);
			}
			break;

		case RepoActions.FetchedRepo:
			if (action.repoObj.IsCloning) {
				this.logEvent("AddRepo", {
					private: Boolean(action.repoObj.Private),
					language: action.repoObj.Language,
				});
			}
			break;

		case UserActions.SignupCompleted:
			this.logEvent("SignupCompleted", {error: Boolean(action.resp.Error)});
			if (!action.resp.Error) {
				this.setUserProperty("email", action.email);
			}
			break;

		case UserActions.LoginCompleted:
		case UserActions.LogoutCompleted:
		case UserActions.ForgotPasswordCompleted:
		case UserActions.ResetPasswordCompleted:
			if (action.eventName) {
				this.logEvent(action.eventName, {error: Boolean(action.resp.Error)});
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

		this._flush(); // No need to __emitChange(); components need not be re-rendered.
	}
}

export default new EventLogger();
