import * as CodeActions from "sourcegraph/code/CodeActions";
import CodeStore from "sourcegraph/code/CodeStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

const CodeBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.WantFile:
			{
				let file = CodeStore.files.get(action.repo, action.rev, action.tree);
				if (file === null) {
					let revPart = action.rev ? `@${action.rev}` : "";
					let uri = `/${action.repo}${revPart}/.tree/${action.tree}`;

					if (typeof window !== "undefined" && window.preloadedCodeViewFile && window.preloadedCodeViewFile.url === uri) {
						Dispatcher.asyncDispatch(new CodeActions.FileFetched(action.repo, action.rev, action.tree, JSON.parse(window.preloadedCodeViewFile.data)));
						return;
					}

					CodeBackend.xhr({
						uri: `/.ui${uri}`,
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
		}
	},
};

Dispatcher.register(CodeBackend.__onDispatch);

export default CodeBackend;
