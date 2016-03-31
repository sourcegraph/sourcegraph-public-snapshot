import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";
import prepareAnnotations from "sourcegraph/blob/prepareAnnotations";

const BlobBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case BlobActions.WantFile:
			{
				let file = BlobStore.files.get(action.repo, action.rev, action.tree);
				if (file === null) {
					let revPart = action.rev ? `@${action.rev}` : "";
					BlobBackend.xhr({
						uri: `/.api/repos/${action.repo}${revPart}/-/tree/${action.tree}?ContentsAsString=true`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.Stores.dispatch(new BlobActions.FileFetched(action.repo, action.rev, action.tree, body));
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
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						body.Annotations = prepareAnnotations(body.Annotations);
						Dispatcher.Stores.dispatch(new BlobActions.AnnotationsFetched(action.repo, action.rev, action.commitID, action.path, action.startByte, action.endByte, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(BlobBackend.__onDispatch);

export default BlobBackend;
