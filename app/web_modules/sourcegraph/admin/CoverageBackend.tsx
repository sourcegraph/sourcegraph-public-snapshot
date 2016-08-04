// tslint:disable

import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import CoverageStore from "sourcegraph/admin/CoverageStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const CoverageBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case CoverageActions.WantCoverage:
			{
				let coverage = CoverageStore.coverage;
				if (coverage === null) {
					CoverageBackend.fetch("/.api/coverage")
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new CoverageActions.CoverageFetched(data)));
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(CoverageBackend.__onDispatch);

export default CoverageBackend;
