import FluxUtils from "flux/utils";

class Store extends FluxUtils.Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.reset();

		// reset store for each test
		if (global.beforeEach) {
			global.beforeEach(() => { this.reset(); });
		}
	}
}

export default Store;
