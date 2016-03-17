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
};
