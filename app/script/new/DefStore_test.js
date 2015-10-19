import expect from "expect.js";

import Dispatcher from "./Dispatcher";
import DefStore from "./DefStore";
import * as DefActions from "./DefActions";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("/someURL", "someData"));
		expect(DefStore.defs.get("/someURL")).to.be("someData");
	});

	it("should handle HighlightDef", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("someDef"));
		expect(DefStore.highlightedDef).to.be("someDef");

		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef(null));
		expect(DefStore.highlightedDef).to.be(null);
	});
});
