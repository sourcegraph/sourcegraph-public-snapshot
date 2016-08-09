// tslint:disable

import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobStore, keyForFile, keyForAnns} from "sourcegraph/blob/BlobStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {prepareAnnotations} from "sourcegraph/blob/prepareAnnotations";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {updateRepoCloning} from "sourcegraph/repo/cloning";

export const BlobBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
		switch (action.constructor) {
		case BlobActions.WantFile:
			{
				let file = BlobStore.files[keyForFile(action.repo, action.commitID, action.path)] || null;
				if (file === null) {
					let url = `/.api/repos/${action.repo}@${action.commitID}/-/tree/${action.path}?ContentsAsString=true`;
					BlobBackend.fetch(url)
						.then(updateRepoCloning(action.repo))
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => {
							if (data.IncludedAnnotations) {
								const anns = data.IncludedAnnotations;
								delete data.IncludedAnnotations;
								if (anns.Annotations) {
									anns.Annotations = prepareAnnotations(anns.Annotations);
								}
								Dispatcher.Stores.dispatch(new BlobActions.AnnotationsFetched(action.repo, action.commitID, action.path, 0, 0, anns));
							}
							Dispatcher.Stores.dispatch(new BlobActions.FileFetched(action.repo, action.commitID, action.path, data));
						});
				}
				break;
			}

		case BlobActions.WantAnnotations:
			{
				let anns = BlobStore.annotations[keyForAnns(action.repo, action.commitID, action.path, action.startByte, action.endByte)] || null;
				if (anns === null) {
					let url = `/.api/annotations?Entry.RepoRev.Repo=${action.repo}&Entry.RepoRev.CommitID=${action.commitID}&Entry.Path=${action.path}&Range.StartByte=${action.startByte || 0}&Range.EndByte=${action.endByte || 0}`;
					BlobBackend.fetch(url)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => {
							if (!data.Error && data.Annotations) data.Annotations = prepareAnnotations(data.Annotations);
							Dispatcher.Stores.dispatch(
								new BlobActions.AnnotationsFetched(
									action.repo, action.commitID, action.path,
									action.startByte, action.endByte, data));
						});
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(BlobBackend.__onDispatch);
