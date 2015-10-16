import * as DefActions from "./DefActions";
import DefStore from "./DefStore";
import Dispatcher from "./Dispatcher";
import defaultXhr from "xhr";

// TODO preloading
const DefBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			let def = DefStore.defs[action.url];
			if (def === undefined) {
				DefBackend.xhr({
					uri: `/ui${action.url}`,
					headers: {
						"X-Definition-Data-Only": "yes",
					},
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					if (!body.Found) {
						console.warn("def not found");
						return;
					}
					Dispatcher.dispatch(new DefActions.DefFetched(action.url, body));
				});
			}
			break;
		}
	},
};

Dispatcher.register(DefBackend.__onDispatch);

export default DefBackend;
