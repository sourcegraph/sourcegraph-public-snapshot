import * as FluxUtils from "flux/utils";
import {Dispatcher} from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import {testOnly} from "sourcegraph/util/testOnly";

export class Store<TPayload> extends FluxUtils.Store<TPayload> {
	// hack: access internal fields of store and dispatcher
	// tslint:disable-next-line: variable-name
	__dispatcher: {
		_startDispatching: (payload: any) => void,
		_stopDispatching: () => void,
	};

	constructor(dispatcher: Dispatcher) {
		super(dispatcher);
		this.reset();

		// reset store for each test
		if (global.beforeEach) {
			global.beforeEach(() => { this.reset(); });
		}
	}

	// directDispatch dispatches an action to a single store. Not affected by
	// catchDispatched. Use only in tests.
	directDispatch(payload: any): void {
		testOnly();

		deepFreeze(payload);

		this.__dispatcher._startDispatching(payload);
		try {
			this.__onDispatch(payload);
		} finally {
			this.__dispatcher._stopDispatching();
		}
	}

	reset(): void {
		// overrride
	}
}
