import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";

function keyFor(repo, rev, tree) {
	return `${repo}#${rev}#${tree}`;
}

let CodeStore = {
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
		let i = CodeStore.listeners.indexOf(listener);
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

export default CodeStore;
