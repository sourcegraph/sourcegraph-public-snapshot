import React from "react";

export class EventLogger {

    updatePropsForUser(identity) {
        if (identity) {
            chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
            chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
            if (identity.gaClientId) {
                chrome.runtime.sendMessage({ type: "setTrackerGAClientId", payload: identity.gaClientId});
            }
        }
    }

    logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties ? : any) {
        if (process.env.NODE_ENV === "test") return;

        eventProperties = eventProperties ? eventProperties : {};
        eventProperties["Platform"] = window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension";

        chrome.runtime.sendMessage({ type: "trackEvent", payload: Object.assign({}, eventProperties, { eventLabel, eventCategory, eventAction }) });
    }
}

export default new EventLogger();
