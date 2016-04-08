// @flow weak

import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import Dispatcher from "sourcegraph/Dispatcher";
import prepareAnnotations from "sourcegraph/blob/prepareAnnotations";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const BlobBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case BlobActions.WantFile:
			{
				let file = BlobStore.files.get(action.repo, action.rev, action.path);
				if (file === null) {
					let url = `/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/tree/${action.path}?ContentsAsString=true`;
					BlobBackend.fetch(url)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: true}))
							.then((data) => Dispatcher.Stores.dispatch(
								new BlobActions.FileFetched(action.repo, action.rev, action.path, data)));
				}
				break;
			}

		case BlobActions.WantAnnotations:
			{
				let anns = BlobStore.annotations.get(action.repo, action.rev, action.commitID, action.path, action.startByte, action.endByte);
				if (anns === null) {
					let url = `/.api/annotations?Entry.RepoRev.URI=${action.repo}&Entry.RepoRev.Rev=${action.rev}&Entry.RepoRev.CommitID=${action.commitID}&Entry.Path=${action.path}&Range.StartByte=${action.startByte || 0}&Range.EndByte=${action.endByte || 0}`;
					BlobBackend.fetch(url)
							.then(checkStatus)
							.then((resp) => resp.json())
							.then((data) => {
								data.Annotations = prepareAnnotations(data.Annotations);
								Dispatcher.Stores.dispatch(
									new BlobActions.AnnotationsFetched(
										action.repo, action.rev, action.commitID, action.path,
										action.startByte, action.endByte, data));
							})
							.catch((err) => console.error(err));
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(BlobBackend.__onDispatch);

export default BlobBackend;
