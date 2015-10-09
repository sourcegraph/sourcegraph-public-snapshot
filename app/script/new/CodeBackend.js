var CodeActions = require("./CodeActions");
var CodeStore = require("./CodeStore");
var Dispatcher = require("./Dispatcher");

// TODO preloading
var CodeBackend = {
	xhr: require("xhr"),

	handle(action) {
		switch (action.constructor) {
		case CodeActions.WantFile:
			var file = CodeStore.files.get(action.repo, action.rev, action.tree);
			if (file === undefined) {
				CodeBackend.xhr({
					uri: `/ui/${action.repo}@${action.rev}/.tree/${action.tree}`,
					json: {},
				}, function(err, resp, body) {
					// TODO handle error
					Dispatcher.dispatch(new CodeActions.SetFile(action.repo, action.rev, action.tree, body));
				});
			}
			break;
		}
	},
};

Dispatcher.register(CodeBackend.handle);

module.exports = CodeBackend;
