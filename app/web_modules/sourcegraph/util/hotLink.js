import Dispatcher from "sourcegraph/Dispatcher";

export class GoTo {
	constructor(url) {
		this.url = url;
	}
}

export default function hotLink(event) {
	if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey || event.button === 1 || event.button === 2) {
		return;
	}
	event.preventDefault();
	Dispatcher.Stores.dispatch(new GoTo(event.currentTarget.href || event.currentTarget.dataset.href));
}

// hotLinkAnyElement lets clicks on child A elements continue with the default
// behavior. Otherwise it behaves like hotLink.
export function hotLinkAnyElement(event) {
	if (!event.target) return null;
	if (event.target.tagName === "A") return null;
	return hotLink(event);
}
