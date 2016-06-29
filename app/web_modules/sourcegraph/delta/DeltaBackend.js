// @flow weak

import * as DeltaActions from "sourcegraph/delta/DeltaActions";
import DeltaStore from "sourcegraph/delta/DeltaStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";

const DeltaBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
		if (action instanceof DeltaActions.WantFiles) {
			let commit = DeltaStore.files.get(action.baseRepo, action.baseRev, action.headRepo, action.headRev);
			if (commit === null) {
				DeltaBackend.fetch(`/.api/repos/${action.headRepo}@${action.headRev}/-/delta/${action.baseRev}/-/files`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new DeltaActions.FetchedFiles(action.baseRepo, action.baseRev, action.headRepo, action.headRev, data));
					});
			}
			return;
		}
	},
};

Dispatcher.Backends.register(DeltaBackend.__onDispatch);

export default DeltaBackend;
