import * as CodeActions from "sourcegraph/code/CodeActions";
import {authHeaders} from "sourcegraph/util/auth";
import {sortAnns} from "sourcegraph/code/Annotations";
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

		case CodeActions.WantAnnotations:
			{
				let anns = CodeStore.annotations.get(action.repo, action.rev, action.path, action.startByte, action.endByte);
				if (anns === null) {
					CodeBackend.xhr({
						uri: `/.api/annotations?Entry.RepoRev.URI=${action.repo}&Entry.RepoRev.Rev=${action.rev}&Entry.Path=${action.path}&Range.StartByte=${action.startByte || 0}&Range.EndByte=${action.endByte || 0}`,
						json: {},
						headers: authHeaders(),
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						body.Annotations = prepareAnnotations(body.Annotations);
						Dispatcher.dispatch(new CodeActions.AnnotationsFetched(action.repo, action.rev, action.path, action.startByte, action.endByte, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(CodeBackend.__onDispatch);

export default CodeBackend;

// prepareAnnotations should be called on annotations received from the server
// to prepare them in ways described below for presentation in the UI.
export function prepareAnnotations(anns) {
	// Ensure that syntax highlighting is the innermost annotation so
	// that the CSS colors are applied (otherwise ref links appear in
	// the normal link color).
	anns.forEach((a) => {
		if (!a.URL) a.WantInner = 1;
	});

	sortAnns(anns);

	// Condense coincident refs ("multiple defs", such as an embedded Go
	// field's ref to both the field def and the type def).
	for (let i = 0; i < anns.length; i++) {
		const ann = anns[i];
		for (let j = i + 1; j < anns.length; j++) {
			const ann2 = anns[j];
			if (ann.StartByte === ann2.StartByte && ann.EndByte === ann2.EndByte) {
				if ((ann.URLs || ann.URL) && ann2.URL) {
					ann.URLs = (ann.URLs || [ann.URL]).concat(ann2.URL);
					delete ann.URL;
					anns.splice(j, 1); // Delete the coincident ref.
				}
			} else {
				break;
			}
		}
	}

	return anns;
}
