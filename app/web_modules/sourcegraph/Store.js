import FluxUtils from "flux/utils";
import deepFreeze from "sourcegraph/util/deepFreeze";
import testOnly from "sourcegraph/util/testOnly";

class Store extends FluxUtils.Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();

		// reset store for each test
		if (global.beforeEach) {
			global.beforeEach(() => { this.reset(); });
		}
	}

	// directDispatch dispatches an action to a single store. Not affected by
	// catchDispatched. Use only in tests.
	directDispatch(payload) {
		testOnly();

		deepFreeze(payload);

		this.__dispatcher._startDispatching(payload);
		try {
			this.__onDispatch(payload);
		} finally {
			this.__dispatcher._stopDispatching();
		}
	}
}

export default Store;
