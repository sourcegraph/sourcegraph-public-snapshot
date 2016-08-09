// tslint:disable

import {testOnly} from "sourcegraph/util/testOnly";

// immediateSyncPromise returns an object that looks like a resolved promise
// but whose then and catch callbacks are executed synchronously
// and immediately.
//
// Only use it in tests.
export function immediateSyncPromise(val, isError?) {
	testOnly();
	return {
		then: (resolve, reject) => {
			if (!resolve && reject) return immediateSyncPromise(val, isError).catch(reject);
			if (isError) return immediateSyncPromise(val, isError);
			try {
				let val2 = resolve(val);
				return (val2 && val2.then && val2.catch) ? val2 : immediateSyncPromise(val2, false);
			} catch (e) {
				let p2 = immediateSyncPromise(e, true);
				if (reject) p2 = p2.catch(reject);
				return p2;
			}
		},
		catch: (reject) => immediateSyncPromise(isError ? reject(val) : val, false),
	};
}
