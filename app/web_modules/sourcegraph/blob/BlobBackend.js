import * as BlobActions from "sourcegraph/blob/BlobActions";
import {authHeaders} from "sourcegraph/util/auth";
import {sortAnns} from "sourcegraph/blob/Annotations";
import BlobStore from "sourcegraph/blob/BlobStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

const BlobBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case BlobActions.WantFile:
			{
				let file = BlobStore.files.get(action.repo, action.rev, action.tree);
				if (file === null) {
					let revPart = action.rev ? `@${action.rev}` : "";
					let uri = `/${action.repo}${revPart}/.tree/${action.tree}`;

					if (typeof window !== "undefined" && window.preloadedBlob && window.preloadedBlob.url === uri) {
						Dispatcher.asyncDispatch(new BlobActions.FileFetched(action.repo, action.rev, action.tree, JSON.parse(window.preloadedBlob.data)));
						return;
					}

					BlobBackend.xhr({
						uri: `/.ui${uri}`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new BlobActions.FileFetched(action.repo, action.rev, action.tree, body));
					});
				}
				break;
			}

		case BlobActions.WantAnnotations:
			{
				let anns = BlobStore.annotations.get(action.repo, action.rev, action.commitID, action.path, action.startByte, action.endByte);
				if (anns === null) {
					BlobBackend.xhr({
						uri: `/.api/annotations?Entry.RepoRev.URI=${action.repo}&Entry.RepoRev.Rev=${action.rev}&Entry.RepoRev.CommitID=${action.commitID}&Entry.Path=${action.path}&Range.StartByte=${action.startByte || 0}&Range.EndByte=${action.endByte || 0}`,
						json: {},
						headers: authHeaders(),
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						body.Annotations = prepareAnnotations(body.Annotations);
						Dispatcher.dispatch(new BlobActions.AnnotationsFetched(action.repo, action.rev, action.commitID, action.path, action.startByte, action.endByte, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(BlobBackend.__onDispatch);

export default BlobBackend;

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
