/**
 * TODO: This module should be rewritten! We shouldn't be using DOM events
 * anymore since we're not communicating across separate JS <-> TS scripts.
 */

/**
 * done marks syntax highlighting as done.
 */
export function done(): void {
	const finishEvent = document.createEvent("Event");
	finishEvent.initEvent("syntaxHighlightingFinished", true, true);
	window.dispatchEvent(finishEvent);
}

let syntaxHighlightingFinished = false;

window.addEventListener("syntaxHighlightingFinished", () => {
	syntaxHighlightingFinished = true;
}, false);

/**
 * wait returns a promise that waits for syntax highlighting to be finished.
 */
export function wait(): Promise<void> {
	if (syntaxHighlightingFinished) {
		return Promise.resolve();
	}
	return new Promise((resolve, _reject) => {
		window.addEventListener("syntaxHighlightingFinished", () => {
			resolve();
		}, false);
	});
}
