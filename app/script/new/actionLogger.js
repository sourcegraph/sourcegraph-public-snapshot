import Dispatcher from "./Dispatcher";

Dispatcher.register(function(action) {
	if (window.localStorage["log-actions"] === "true") {
		console.log(action);
	}
});

window.enableActionLog = function() {
	window.localStorage["log-actions"] = "true";
	console.log("Action log enabled.");
};

window.disableActionLog = function() {
	Reflect.deleteProperty(window.localStorage, "log-actions");
	console.log("Action log disabled.");
};

console.log(`Welcome to JS console. Action log is ${window.localStorage["log-actions"] === "true" ? "enabled" : "disabled"}. Use enableActionLog() and disableActionLog() to change.`);
