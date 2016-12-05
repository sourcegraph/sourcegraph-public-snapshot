import * as flux from "flux";

import { deepFreeze } from "sourcegraph/util/deepFreeze";
import { testOnly } from "sourcegraph/util/testOnly";

export class Dispatcher extends flux.Dispatcher<any> {
	_catch: boolean;
	_dispatched: any[];

	dispatch(payload: any): void {
		deepFreeze(payload);
		if (this._catch) {
			this._dispatched.push(payload);
			return;
		}
		super.dispatch(payload);
	}

	// catchDispatched returns the actions dispatched during the execution of f.
	// Actions are not passed to the listeners. Use only in tests.
	catchDispatched(f: () => void): any[] {
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

export const Stores = new Dispatcher();
export const Backends = new Dispatcher();
