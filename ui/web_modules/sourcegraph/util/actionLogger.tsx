import * as Dispatcher from "sourcegraph/Dispatcher";
import { Features } from "sourcegraph/util/features";

if (typeof window !== "undefined") {
	let logger = (dispatcherName) => (action) => {
		if (Features.actionLogDebug.isEnabled()) {
			// tslint:disable-next-line
			console.log(`${dispatcherName}:`, action);
		}
	};

	Dispatcher.Stores.register(logger("Stores"));
	Dispatcher.Backends.register(logger("Backends"));
}
