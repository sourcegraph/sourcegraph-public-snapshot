import { telligent } from "sourcegraph/tracking/services/telligentWrapper";
import { getPathExtension } from "sourcegraph/util";
import { pageVars } from "sourcegraph/util/pageVars";
import { sourcegraphContext } from "sourcegraph/util/sourcegraphContext";
import * as url from "sourcegraph/util/url";

class EventLogger {
	private static PLATFORM: string = "Web";

	constructor() {
		this.updateUser();
		this.updateTrackerWithIdentificationProps();

		// TODO(Dan): validate that this communication is working and sufficient
		// Because the new webapp isn't a single page app, we send a 2nd message after 1000ms,
		// in case the injected DOM element wasn't available in time for the first call
		setTimeout(() => this.updateTrackerWithIdentificationProps(), 1000);
	}

	/**
	 * Set user-level properties in all external tracking services
	 */
	updateUser(): void {
		const user = sourcegraphContext.user;
		if (user) {
			this.setUserId(user.UID.toString(), user.Login);
		}

		const email = sourcegraphContext.primaryEmail();
		if (email) {
			this.setUserEmail(email);
		}
	}

	/**
	 * Function to sync our key user identification props across Telligent and user Chrome extension installations
	 */
	updateTrackerWithIdentificationProps(): any {
		if (!telligent.isTelligentLoaded() || !sourcegraphContext.hasBrowserExtensionInstalled()) {
			return null;
		}

		this.setUserInstalledChromeExtension("true");

		const idProps = { detail: { deviceId: telligent.getTelligentDuid(), userId: sourcegraphContext.user && sourcegraphContext.user.Login } };
		setTimeout(() => document.dispatchEvent(new CustomEvent("sourcegraph:identify", idProps)), 20);
	}

	setUserId(UID: string, login: string): void {
		telligent.setUserId(login);
		telligent.setUserProperty("internal_user_id", UID);
	}

	setUserRegisteredAt(registeredAt: any): void {
		telligent.setUserProperty("registered_at_timestamp", registeredAt);
		telligent.setUserProperty("registered_at", new Date(registeredAt).toDateString());
	}

	setUserInstalledChromeExtension(installedChromeExtension: string): void {
		telligent.setUserProperty("installed_chrome_extension", installedChromeExtension);
	}

	setUserEmail(primaryEmail: string): void {
		telligent.setUserProperty("email", primaryEmail);
	}

	setUserInvited(invitingUserId: string, invitedToOrg: string): void {
		telligent.setUserProperty("invited_by_user", invitingUserId);
		telligent.setUserProperty("org_invite", invitedToOrg);
	}

	/**
	 * Tracking call to analytics services on pageview events
	 * Note: should NEVER be called outside of events.tsx
	 */
	logViewEvent(pageTitle: string, eventProperties?: any): void {
		if (sourcegraphContext.userAgentIsBot || !pageTitle) {
			return;
		}

		const decoratedProps = { ...this.decorateEventProperties(eventProperties), page_name: pageTitle, page_title: pageTitle };
		telligent.track("view", decoratedProps);
		this.logToConsole(pageTitle, decoratedProps);
	}

	/**
	 * Tracking call to analytics services on user action events
	 * Note: should NEVER be called outside of events.tsx
	 */
	logEvent(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any): void {
		if (sourcegraphContext.userAgentIsBot || !eventLabel) {
			return;
		}

		const decoratedProps = { ...this.decorateEventProperties(eventProperties), eventLabel: eventLabel, eventCategory: eventCategory, eventAction: eventAction };
		telligent.track(eventAction, decoratedProps);
		this.logToConsole(eventLabel, decoratedProps);
	}

	private logToConsole(eventLabel: string, object?: any): void {
		if (localStorage && localStorage.getItem("eventLogDebug") === "true") {
			console.debug("%cEVENT %s", "color: #aaa", eventLabel, object); // tslint:disable-line
		}
	}

	private decorateEventProperties(platformProperties: any): any {
		const props = {
			...platformProperties,
			platform: EventLogger.PLATFORM,
			is_authed: sourcegraphContext.user ? "true" : "false",
			path_name: window.location && window.location.pathname ? window.location.pathname.slice(1) : "",
		};

		const u = url.parseBlob();
		if (u.uri) {
			props.repo = u.uri!;
			props.rev = pageVars.ResolvedRev;
			if (u.path) {
				props.path = u.path!;
				props.language = getPathExtension(u.path);
			}
		}

		return props;
	}
}

export const eventLogger = new EventLogger();
