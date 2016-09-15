import * as BlobActions from "sourcegraph/blob/BlobActions";
import { BlobStore, keyForFile } from "sourcegraph/blob/BlobStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { updateRepoCloning } from "sourcegraph/repo/cloning";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

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
					.catch((err) => ({ Error: err }))
					.then((data) => {
						Dispatcher.Stores.dispatch(new BlobActions.FileFetched(action.repo, action.commitID, action.path, data));
					});
			}
		}
	},
};

Dispatcher.Backends.register(BlobBackend.__onDispatch);
