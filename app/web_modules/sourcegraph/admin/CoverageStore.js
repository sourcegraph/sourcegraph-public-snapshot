import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as CoverageActions from "sourcegraph/admin/CoverageActions";
import "sourcegraph/admin/CoverageBackend";

export class CoverageStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	toJSON() {
		return {
			coverage: this.coverage,
		};
	}

	reset(data) {
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

export default new CoverageStore(Dispatcher.Stores);
