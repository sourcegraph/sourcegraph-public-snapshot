import expect from "expect.js";

import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", "someData"));
		expect(DefStore.defs.get("r", "v", "d")).to.be("someData");
	});

	it("should handle DefsFetched", () => {
		DefStore.directDispatch(new DefActions.DefsFetched("r", "v", "q", ["someData"]));
		expect(DefStore.defs.list("r", "v", "q")).to.eql(["someData"]);
	});

	it("should handle HighlightDef", () => {
		DefStore.directDispatch(new DefActions.HighlightDef("someDef"));
		expect(DefStore.highlightedDef).to.be("someDef");

		DefStore.directDispatch(new DefActions.HighlightDef(null));
		expect(DefStore.highlightedDef).to.be(null);
	});

	it("should handle RefLocationsFetched", () => {
		DefStore.directDispatch(new DefActions.RefLocationsFetched("r", "v", "d", ["someData"]));
		expect(DefStore.refLocations.get("r", "v", "d")).to.eql(["someData"]);
	});

	it("should handle RefsFetched", () => {
		DefStore.directDispatch(new DefActions.RefsFetched("r", "v", "d", "rr", "rf", ["someData"]));
		expect(DefStore.refs.get("r", "v", "d", "rr", "rf")).to.eql(["someData"]);
	});
});
