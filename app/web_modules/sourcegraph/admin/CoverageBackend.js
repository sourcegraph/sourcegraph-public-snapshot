import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import CoverageStore from "sourcegraph/admin/CoverageStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";

const CoverageBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case CoverageActions.WantCoverage:
			{
				let coverage = CoverageStore.coverage;
				if (coverage === null) {
					trackPromise(
						CoverageBackend.fetch("/.api/coverage")
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new CoverageActions.CoverageFetched(data)))
					);
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(CoverageBackend.__onDispatch);

export default CoverageBackend;
