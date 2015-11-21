import * as CodeActions from "./CodeActions";
import CodeStore from "./CodeStore";
import Dispatcher from "../Dispatcher";
import defaultXhr from "xhr";

// TODO preloading
const CodeBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.WantFile:
			let file = CodeStore.files.get(action.repo, action.rev, action.tree);
			if (file === null) {
				CodeBackend.xhr({
					uri: `/.ui/${action.repo}@${action.rev}/.tree/${action.tree}`,
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					Dispatcher.dispatch(new CodeActions.FileFetched(action.repo, action.rev, action.tree, body));
				});
			}
			break;
		}
	},
};

Dispatcher.register(CodeBackend.__onDispatch);

export default CodeBackend;
