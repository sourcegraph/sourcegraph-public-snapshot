// This file manages event logging functionality for definfo and repoinfo landing pages,
// which do not load the full ui tree

// Global variable
var sourcegraphLogger;

(function() { 

	function StaticEventLoggerClass() {
		if (!window) return;

		this._jsContext = window.__sourcegraphJSContext;
		this._telligent = window.telligent;
		this._ga = window.ga;
		this._gaClientID = null;

		let env = "development";
		if (this._jsContext.buildVars.Version !== "dev") {
			switch (this._jsContext.appURL) {
				case "https://sourcegraph.com":
					env = "production";
					break;
				default:
					break;
			}
		}

		// Initialize telligent tracker
		this._telligent("newTracker", "sgDefInfoTracker", "sourcegraph-logging.telligentdata.com", {
			appId: "SourcegraphWeb",
			platform: "Web",
			encodeBase64: false,
			env: env,
			configUseCookies: true,
			useCookies: true,
			metadata: {
				gaCookies: true,
				webPage: true,
			},
		});

		// Set user id if logged in 
		if (this._jsContext.user) {
			if (this._ga) {
				this._ga("set", "userId", this._jsContext.user.Login);
			}

			if (this._telligent) {
				this._telligent("setUserId", this._jsContext.user.Login);
			}

			this._setUserProperty("internal_user_id", this._jsContext.user.UID.toString());
		}

		// Send gaClientId prop to telligent (fallback for user identification if not logged in) 
		if (this._ga) {
			// Get GA clientId	
			this._ga(function(tracker) {
				this._gaClientID = tracker.get("clientId");
				this._telligent("addStaticMetadataObject", {deviceInfo: {GAClientId: this._gaClientID}});
			}.bind(this));
		}
	};
		
	StaticEventLoggerClass.prototype = {
		// sets current user's properties
		_setUserProperty: function(property, value) {
			if (this._telligent) {
				this._telligent("addStaticMetadata", property, value, "userInfo");
			}
		},

		_decorateEventProperties: function(platformProperties) {
			return Object.assign({}, platformProperties, {Platform: "Web", platformVersion: "", is_authed: this._jsContext.user ? "true" : "false", path_name: window && window.location && window.location.pathname ? window.location.pathname.slice(1) : ""});
		},

		// Use logViewEvent as the default way to log view events for Telligent and GA
		// location is the URL, page is the path.
		logViewEvent: function(title, page, eventProperties) {
			if (this._jsContext.userAgentIsBot || !page) {
				return;
			}

			if (this._telligent) {
				this._logToConsole(title, Object.assign({}, this._decorateEventProperties(eventProperties), {page_name: page, page_title: title}));
				this._telligent("track", "view", Object.assign({}, this._decorateEventProperties(eventProperties), {page_name: page, page_title: title}));
			}
		},
		logDefLandingViewEvent: function() {
			this.logViewEvent("ViewDefLanding", location.pathname, {});
		},
		logRepoIndexViewEvent: function() {
			this.logViewEvent("ViewRepoIndex", location.pathname, {});
		},

		// Default tracking call to all of our analytics servies.
		// By default, should only be called by AnalyticsConstants.LoggableEvent.logEvent()
		// Required fields: event
		// Optional fields: eventProperties
		logEventForCategory: function(event, eventProperties) {
			this.logEventForCategoryComponents(event.category, event.action, event.label, eventProperties);
		},

		logEventForCategoryComponents: function(eventCategory, eventAction, eventLabel, eventProperties) {
			if (this._jsContext.userAgentIsBot || !eventLabel) {
				return;
			}
			if (this._telligent) {
				this._telligent("track", eventAction, Object.assign({}, this._decorateEventProperties(eventProperties), {eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction}));
			}

			this._logToConsole(eventAction, Object.assign(this._decorateEventProperties(eventProperties),  {eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction}));

			if (this._ga) {
				this._ga("send", {
					hitType: "event",
					eventCategory: eventCategory || "",
					eventAction: eventAction || "",
					eventLabel: eventLabel,
				});
			}
		},

		_logToConsole: function(eventAction, object) {
			if (window && window.localStorage && window.localStorage["log_debug"]) {
				console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
			}
		}
	};

	sourcegraphLogger = new StaticEventLoggerClass();
	
	// Log view events on load
	var head_elt = document.getElementsByTagName("HEAD");
	if (head_elt && head_elt[0]) {
		var page_title = head_elt[0].getAttribute("data-template-name");
		if (page_title) {
			switch (page_title) {
				case "deflanding.html": sourcegraphLogger.logDefLandingViewEvent(); break;
				case "repoindex.html": sourcegraphLogger.logRepoIndexViewEvent(); break;
			}
		}
	}
	
})();

