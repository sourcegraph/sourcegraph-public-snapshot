/* tslint:disable: no-console */
import * as Dispatcher from "sourcegraph/Dispatcher";

if (typeof window !== "undefined") {
	let logger = function(dispatcherName: string): (action: any) => void {
		return function(action: any): void {
			if (window.localStorage["log-actions"] === "true") {
				console.log(`${dispatcherName}:`, action);
			}
		};
	};

	Dispatcher.Stores.register(logger("Stores"));
	Dispatcher.Backends.register(logger("Backends"));

	(window as any).enableActionLog = function(): void {
		window.localStorage["log-actions"] = "true";
		console.log("Action log enabled.");
	};

	(window as any).disableActionLog = function(): void {
		Reflect.deleteProperty(window.localStorage, "log-actions");
		console.log("Action log disabled.");
	};

	console.log(`Welcome to JS console. Action log is ${window.localStorage["log-actions"] === "true" ? "enabled" : "disabled"}. Use enableActionLog() and disableActionLog() to change.`);
}
