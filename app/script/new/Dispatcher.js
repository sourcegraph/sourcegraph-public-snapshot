import flux from "flux";

import testOnly from "./util/testOnly";

class Dispatcher extends flux.Dispatcher {
	dispatch(payload) {
		if (this._catch) {
			this._dispatched.push(payload);
			return;
		}
		super.dispatch(payload);
	}

	asyncDispatch(payload) {
		setTimeout(() => {
			this.dispatch(payload);
		}, 0);
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

	// directDispatch dispatches an action to a single store. Not affected by
	// catchDispatched. Use only in tests.
	directDispatch(store, payload) {
		testOnly();

		this._startDispatching(payload);
		try {
			store.__onDispatch(payload);
		} finally {
			this._stopDispatching();
		}
	}
}

export default new Dispatcher();
