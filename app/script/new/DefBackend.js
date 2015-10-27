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
			let def = DefStore.defs.get(action.url);
			if (def === null) {
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

		case DefActions.WantExample:
			let example = DefStore.examples.get(action.defURL, action.index);
			if (example === null) {
				DefBackend.xhr({
					uri: `/ui${action.defURL}/.examples?TokenizedSource=true&PerPage=1&Page=${action.index + 1}`,
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					if (body === null) {
						Dispatcher.dispatch(new DefActions.NoExampleAvailable(action.defURL, action.index));
						return;
					}
					Dispatcher.dispatch(new DefActions.ExampleFetched(action.defURL, action.index, body[0]));
				});
			}
			break;

		}
	},
};

Dispatcher.register(DefBackend.__onDispatch);

export default DefBackend;
