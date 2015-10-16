import expect from "expect.js";

import Dispatcher from "./Dispatcher";
import DefStore from "./DefStore";
import * as DefActions from "./DefActions";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("/someURL", "someData"));
		expect(DefStore.defs["/someURL"]).to.be("someData");
	});
});
