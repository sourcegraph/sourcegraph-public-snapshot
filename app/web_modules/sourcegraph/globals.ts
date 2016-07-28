/* tslint:disable: no-namespace */

namespace process {
	export var env: {NODE_ENV: string};
}

namespace global {
	export var it: any; // only set while testing
	export var beforeEach: (f: () => void) => void;
}

declare module "flux/lib/FluxStoreGroup" {
	class FluxStoreGroup {
		constructor(stores: FluxUtils.Store<any>[], callback: () => void);
		release(): void;
	}
	export default FluxStoreGroup;
}
