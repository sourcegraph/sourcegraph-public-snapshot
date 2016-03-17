import Dispatcher from "sourcegraph/Dispatcher";

if (typeof window !== "undefined") {
	let logger = function(dispatcherName) {
		return function(action) {
			if (window.localStorage["log-actions"] === "true") {
				console.log(`${dispatcherName}:`, action);
			}
		};
	};

	Dispatcher.Stores.register(logger("Stores"));
	Dispatcher.Backends.register(logger("Backends"));

	window.enableActionLog = function() {
		window.localStorage["log-actions"] = "true";
		console.log("Action log enabled.");
	};

	window.disableActionLog = function() {
		Reflect.deleteProperty(window.localStorage, "log-actions");
		console.log("Action log disabled.");
	};

	console.log(`Welcome to JS console. Action log is ${window.localStorage["log-actions"] === "true" ? "enabled" : "disabled"}. Use enableActionLog() and disableActionLog() to change.`);
}
