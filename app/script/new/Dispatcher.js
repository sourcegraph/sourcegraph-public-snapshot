import flux from "flux";

import testOnly from "./testOnly";

class Dispatcher extends flux.Dispatcher {
	dispatch(payload) {
		if (this._catch) {
			this._dispatched.push(payload);
			return;
		}
		super.dispatch(payload);
	}

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
