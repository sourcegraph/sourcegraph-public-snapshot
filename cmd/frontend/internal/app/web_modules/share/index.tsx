import * as copy from 'copy-to-clipboard';
import { events } from 'app/tracking/events';

export function injectShareWidget() {
	const a = document.querySelector(".navbar .share") as HTMLAnchorElement;
	if (!a) {
		return; // probably on some other page that doesn't have a share button.
	}
	a.addEventListener("click", (e) => {
		events.ShareButtonClicked.log();

		e.preventDefault();
		if (!a.hasAttribute("href")) {
			return;
		}

		copy(window.location.href);
		const textSpan = a.querySelector(".text");
		const oldText = textSpan!.innerHTML;
		a.removeAttribute("href"); // make it non-clickable
		textSpan!.innerHTML = "Copied link to clipboard!";
		setTimeout(() => {
			a.setAttribute("href", "#"); // make it clickable
			textSpan!.innerHTML = oldText;
		}, 3000);
	});
}
