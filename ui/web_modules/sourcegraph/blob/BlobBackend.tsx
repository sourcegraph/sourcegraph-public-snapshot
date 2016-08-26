import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobStore, keyForAnns, keyForFile} from "sourcegraph/blob/BlobStore";
import {prepareAnnotations} from "sourcegraph/blob/prepareAnnotations";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

export const BlobBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(payload: BlobActions.Action): void {
		if (payload instanceof BlobActions.WantFile) {
			const action = payload;
			let file = BlobStore.files[keyForFile(action.repo, action.commitID, action.path)] || null;
			if (file === null) {
				let url = `/.api/repos/${action.repo}@${action.commitID}/-/tree/${action.path}?ContentsAsString=true&NoSrclibAnns=true`;
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
		}

		if (payload instanceof BlobActions.WantAnnotations) {
			const action = payload;
			let anns = BlobStore.annotations[keyForAnns(action.repo, action.commitID, action.path, action.startByte, action.endByte)] || null;
			if (anns === null) {
				let url = `/.api/annotations?Entry.RepoRev.Repo=${action.repo}&Entry.RepoRev.CommitID=${action.commitID}&Entry.Path=${action.path}&Range.StartByte=${action.startByte || 0}&Range.EndByte=${action.endByte || 0}&NoSrclibAnns=true`;
				BlobBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						if (!data.Error && data.Annotations) {
							data.Annotations = prepareAnnotations(data.Annotations);
						}
						Dispatcher.Stores.dispatch(
							new BlobActions.AnnotationsFetched(
								action.repo, action.commitID, action.path,
								action.startByte, action.endByte, data));
					});
			}
		}
	},
};

Dispatcher.Backends.register(BlobBackend.__onDispatch);
