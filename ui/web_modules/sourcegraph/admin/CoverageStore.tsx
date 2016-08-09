// tslint:disable

import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import "sourcegraph/admin/CoverageBackend";

class CoverageStoreClass extends Store<any> {
	coverage: any;

	constructor(dispatcher) {
		super(dispatcher);
	}

	reset() {
		this.coverage = null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case CoverageActions.CoverageFetched:
			if (action.data && !action.data.Error) {
				const cvgData = action.data.RepoStatuses.map((status) => JSON.parse(status.Description));
				this.coverage = deepFreeze([].concat.apply([], cvgData)); // flatten array of arrays
			} else {
				this.coverage = deepFreeze(action.data);
			}
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const CoverageStore = new CoverageStoreClass(Dispatcher.Stores);
