// Reactivating a change list does not require an additional round trip to the Optimize server.
// As such, a simple and effective approach to Activation events is to fire an event after anything on the page changes.
export function activateDefaultExperiments(): void {
	if (window["dataLayer"]) {
		window["dataLayer"].push({ "event": "optimize.activate" });
	}
}
