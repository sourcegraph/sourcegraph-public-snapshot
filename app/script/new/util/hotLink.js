import Dispatcher from "../Dispatcher";

export class GoTo {
	constructor(url) {
		this.url = url;
	}
}

export default function(event) {
	if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) {
		return;
	}
	event.preventDefault();
	Dispatcher.dispatch(new GoTo(event.target.href));
}
