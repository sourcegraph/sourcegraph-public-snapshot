import flux from "flux";

import deepFreeze from "sourcegraph/util/deepFreeze";
import testOnly from "sourcegraph/util/testOnly";

class Dispatcher extends flux.Dispatcher {
	dispatch(payload) {
		deepFreeze(payload);
		if (this._catch) {
			this._dispatched.push(payload);
			return;
		}
		super.dispatch(payload);
	}

	// catchDispatched returns the actions dispatched during the execution of f.
	// Actions are not passed to the listeners. Use only in tests.
	catchDispatched(f) {
		testOnly();

		this._dispatched = [];
		this._catch = true;
		try {
			f();
		} finally {
			this._catch = false;
		}
		return this._dispatched;
	}
}

export default {
	Stores: new Dispatcher(),
	Backends: new Dispatcher(),

	// directDispatch dispatches an action to a single store. Not affected by
	// catchDispatched. Use only in tests.
	directDispatch(store, payload) {
		testOnly();

		deepFreeze(payload);
		if (store.__dispatcher) {
			store.__dispatcher._startDispatching(payload);
		}
		try {
			store.__onDispatch(payload);
		} finally {
			if (store.__dispatcher) {
				store.__dispatcher._stopDispatching();
			}
		}
	},
};
