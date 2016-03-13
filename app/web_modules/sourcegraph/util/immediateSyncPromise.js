import testOnly from "sourcegraph/util/testOnly";

// immediateSyncPromise returns an object that looks like a resolved promise
// but whose then and catch callbacks are executed synchronously
// and immediately.
//
// Only use it in tests.
export default function immediateSyncPromise(val, isError) {
	testOnly();
	return {
		then: (resolve, reject) => {
			if (!resolve && reject) return immediateSyncPromise(val, isError).catch(reject);
			try {
				let val2 = resolve(val);
				return immediateSyncPromise(val2);
			} catch (e) {
				let p2 = immediateSyncPromise(e, true);
				if (reject) p2 = p2.catch(reject);
				return p2;
			}
		},
		catch: (reject) => immediateSyncPromise(isError ? reject(val) : val, isError),
	};
}
