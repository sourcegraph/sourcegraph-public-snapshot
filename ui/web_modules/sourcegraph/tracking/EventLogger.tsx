import { EmailAddr } from "sourcegraph/api/index";
import { context } from "sourcegraph/app/context";
import { googleAnalytics } from "sourcegraph/tracking/GoogleAnalyticsWrapper";
import { hubSpot } from "sourcegraph/tracking/HubSpotWrapper";
import { intercom } from "sourcegraph/tracking/IntercomWrapper";
import { optimizely } from "sourcegraph/tracking/OptimizelyWrapper";
import { telligent } from "sourcegraph/tracking/TelligentWrapper";
import { experimentManager } from "sourcegraph/util/ExperimentManager";
import { Features } from "sourcegraph/util/features";

class EventLoggerClass {
	private CLOUD_TRACKING_APP_ID: string = "SourcegraphWeb";
	private PLATFORM: string = "Web";

	// init initializes Telligent and Intercom.
	init(): void {
		this.updateUser();
	}

	// updateUser is be called whenever the user changes (on the initial page load).
	//
	// If any events have been buffered, it will flush them immediately.
	// If you do not call updateUser or it is run on the server,
	// any subequent calls to logEvent or setUserProperty will be buffered.
	updateUser(): void {
		const user = context.user;
		const emails = context.emails && context.emails.EmailAddrs || null;
		const primaryEmail = (emails && emails.filter(e => e.Primary).map(e => e.Email)[0]) || null;

		if (user) {
			this.setUserId(user.UID.toString(), user.Login);
		}
		if (context.intercomHash) {
			this.setUserHash(context.intercomHash);
		}

		intercom.boot(context.trackingAppID !== this.CLOUD_TRACKING_APP_ID, context.trackingAppID);

		if (user) {
			if (user.Name) { this.setUserName(user.Name); }
			if (user.RegisteredAt) { this.setUserRegisteredAt(user.RegisteredAt); }
			if (user.Company) { this.setUserCompany(user.Company); }
			if (user.Location) { this.setUserLocation(user.Location); }
			this.setUserIsPrivateCodeUser(context.hasPrivateGitHubToken() ? "true" : "false");
			this.setUserIsGitHubOrgAuthed(context.hasOrganizationGitHubToken() ? "true" : "false");
		}

		if (primaryEmail) {
			this.setUserEmail(primaryEmail, emails);
		}
	}

	logout(): void {
		// Prevent the next user who logs in (e.g., on a public terminal) from
		// seeing the previous user's Intercom messages.
		intercom.shutdown();

		optimizely.logout();
	}

	setUserId(UID: string, login: string): void {
		// Set login name (i.e., GitHub user ID)
		googleAnalytics.setTrackerLogin(login);
		telligent.setUserId(login);
		intercom.setIntercomProperty("business_user_id", login);
		hubSpot.setHubSpotProperties({ "user_id": login });
		optimizely.setUserAttributes({ "user_id": login });

		// Set UID
		intercom.setIntercomProperty("user_id", UID);
		intercom.setIntercomProperty("internal_user_id", UID);
	}

	setUserHash(hash: string): void {
		intercom.setIntercomProperty("user_hash", hash);
		telligent.setUserProperty("user_hash", hash);
	}

	setUserName(name: string): void {
		intercom.setIntercomProperty("name", name);
		telligent.setUserProperty("display_name", name);
		hubSpot.setHubSpotProperties({ "fullname": name });
	}

	setUserCompany(company: string): void {
		telligent.setUserProperty("company", company);
		intercom.setIntercomProperty("company", company);
		hubSpot.setHubSpotProperties({ "company": company });
	}

	setUserRegisteredAt(registeredAt: any): void {
		telligent.setUserProperty("registered_at_timestamp", registeredAt);
		telligent.setUserProperty("registered_at", new Date(registeredAt).toDateString());
		intercom.setIntercomProperty("created_at", new Date(registeredAt).getTime() / 1000);
		hubSpot.setHubSpotProperties({ "registered_at": new Date(registeredAt).toDateString() });
	}

	setUserLocation(location: string): void {
		telligent.setUserProperty("location", location);
		hubSpot.setHubSpotProperties({ "location": location });
	}

	setUserIsPrivateCodeUser(isPrivateCodeUser: string): void {
		telligent.setUserProperty("is_private_code_user", isPrivateCodeUser);
		hubSpot.setHubSpotProperties({ "is_private_code_user": isPrivateCodeUser });
	}

	setUserInstalledChromeExtension(installedChromeExtension: string): void {
		telligent.setUserProperty("installed_chrome_extension", installedChromeExtension);
	}

	setUserIsGitHubOrgAuthed(isGitHubOrgAuthed: string): void {
		telligent.setUserProperty("is_github_organization_authed", isGitHubOrgAuthed);
	}

	setUserEmail(primaryEmail: string, emails?: EmailAddr[] | null): void {
		telligent.setUserProperty("email", primaryEmail);
		telligent.setUserProperty("emails", emails);
		intercom.setIntercomProperty("email", primaryEmail);
		optimizely.setUserAttributes({ "email": primaryEmail });
		hubSpot.setHubSpotProperties({ "email": primaryEmail });
		hubSpot.setHubSpotProperties({ "emails": emails ? emails.map(email => { return email.Email; }).join(",") : "" });
	}

	setUserGitHubAuthedLanguages(languages: string[]): void {
		telligent.setUserProperty("authed_languages_github", languages);
	}

	setUserNumRepos(numRepos: number): void {
		telligent.setUserProperty("num_repos_github", numRepos);
	}

	setUserGitHubAuthedOrgs(orgNames: string[]): void {
		telligent.setUserProperty("authed_orgs_github", orgNames);
		intercom.setIntercomProperty("authed_orgs_github", orgNames);
		hubSpot.setHubSpotProperties({ "authed_orgs_github": orgNames.join(",") });
	}

	setUserIsEmployee(isEmployee: boolean): void {
		telligent.setUserProperty("is_employee", isEmployee);
		optimizely.setUserAttributes({ "is_employee": isEmployee });
	}

	setUserIsGitHubAuthed(isGitHubAuthed: string): void {
		telligent.setUserProperty("github_authed", isGitHubAuthed);
	}

	setUserInvited(invitingUserId: string, invitedToOrg: string): void {
		telligent.setUserProperty("invited_by_user", invitingUserId);
		telligent.setUserProperty("org_invite", invitedToOrg);
	}

	/*
	* Function to extract the Telligent user ID from the first-party cookie set by the Telligent JavaScript Tracker
	*
	* @return string or bool The ID string if the cookie exists or false if the cookie has not been set yet
	*/
	_getTelligentDuid(): string | null {
		let cookieName = "_te_";
		let matcher = new RegExp(cookieName + "id\\.[a-f0-9]+=([^;]+);?");
		let match = document.cookie.match(matcher);
		if (match && match[1]) {
			return match[1].split(".")[0];
		} else {
			return null;
		}
	}

	// Function to sync our key user identification props across Telligent, GA, and user Chrome extension installations
	updateTrackerWithIdentificationProps(): any {
		if (!telligent.isTelligentLoaded() || !context.hasChromeExtensionInstalled()) {
			return null;
		}

		let idProps = { detail: { deviceId: this._getTelligentDuid(), userId: context.user && context.user.Login } };
		if (googleAnalytics.gaClientID) {
			telligent.addStaticMetadataObject({ deviceInfo: { GAClientId: googleAnalytics.gaClientID } });
			setTimeout(() => document.dispatchEvent(new CustomEvent("sourcegraph:identify", Object.assign(idProps, { gaClientId: googleAnalytics.gaClientID }))), 20);
		} else {
			setTimeout(() => document.dispatchEvent(new CustomEvent("sourcegraph:identify", idProps)), 20);
		}
	}

	// Use logViewEvent as the default way to log view events for Telligent and GA
	// location is the URL, page is the path.
	logViewEvent(title: string, page: string, eventProperties: any): void {
		if (context.userAgentIsBot || !page) {
			return;
		}

		this._logToConsole(title, Object.assign({}, this._decorateEventProperties(eventProperties), { page_name: page, page_title: title }));

		telligent.track("view", Object.assign({}, this._decorateEventProperties(eventProperties), { page_name: page, page_title: title }));
	}

	// Tracking call to all of our analytics servies.
	// By default, should only be called by AnalyticsConstants.LoggableEvent.logEvent()
	// Required fields: eventCategory, eventAction, eventLabel
	// Optional fields: eventProperties
	logEventWithComponents(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any): void {
		if (context.userAgentIsBot || !eventLabel) {
			return;
		}
		telligent.track(eventAction, Object.assign({}, this._decorateEventProperties(eventProperties), { eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction }));

		this._logToConsole(eventAction, Object.assign(this._decorateEventProperties(eventProperties), { eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction }));

		experimentManager.logEvent(eventLabel);
		hubSpot.logHubSpotEvent(eventLabel);
		googleAnalytics.logEventCategoryComponents(eventCategory, eventAction, eventLabel);
	}

	// Tracking call for event level calls that we wish to track, but do not wish to impact bounce rate on our site for Google analytics.
	// An example of this would be the event that gets fired following a view event on a Repo that 404s. We fire a view event and then a 404 event.
	// By adding a non-interactive flag to the 404 event the page will correctly calculate bounce rate even with the additional event fired.
	logNonInteractionEventWithComponents(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any): void {
		if (context.userAgentIsBot || !eventLabel) {
			return;
		}

		telligent.track(eventAction, Object.assign({}, this._decorateEventProperties(eventProperties), { eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction }));
		googleAnalytics.logEventCategoryComponents(eventCategory, eventAction, eventLabel, true);

		this._logToConsole(eventAction, Object.assign(this._decorateEventProperties(eventProperties), { eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction, nonInteraction: true }));
	}

	_logToConsole(eventAction: string, object?: any): void {
		if (Features.eventLogDebug.isEnabled()) {
			console.debug("%cEVENT %s", "color: #aaa", eventAction, object); // tslint:disable-line
		}
	}

	_decorateEventProperties(platformProperties: any): any {
		const optimizelyMetadata = optimizely.getOptimizelyMetadata();
		const addtlPlatformProperties = {
			Platform: this.PLATFORM,
			is_authed: context.user ? "true" : "false",
			path_name: global.window && global.window.location && global.window.location.pathname ? global.window.location.pathname.slice(1) : ""
		};
		return Object.assign({}, platformProperties, addtlPlatformProperties, optimizelyMetadata);
	}
}

export const EventLogger = new EventLoggerClass();
