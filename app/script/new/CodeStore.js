var Dispatcher = require("./Dispatcher");
var CodeActions = require("./CodeActions");

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

var CodeStore = {
	files: {
		content: {},
		get(repo, rev, tree) {
			return this.content[keyFor(repo, rev, tree)];
		},
	},
	listeners: [],

	addListener(listener) {
		CodeStore.listeners.push(listener);
	},

	removeListener(listener) {
		var i = CodeStore.listeners.indexOf(listener);
		CodeStore.listeners.splice(i, 1);
	},

	handle(action) {
		switch (action.constructor) {
		case CodeActions.FileFetched:
			CodeStore.files.content[keyFor(action.repo, action.rev, action.tree)] = action.file;
			fireListeners();
		}
	},
};

function fireListeners() {
	CodeStore.listeners.forEach(l => l());
}

Dispatcher.register(CodeStore.handle);

module.exports = CodeStore;
