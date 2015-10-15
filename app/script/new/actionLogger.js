import Dispatcher from "./Dispatcher";

Dispatcher.register(function(action) {
	if (window.localStorage["log-actions"] === "true") {
		console.log(action);
	}
});
