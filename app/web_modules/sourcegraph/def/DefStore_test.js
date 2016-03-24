import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("/someURL", "someData"));
		expect(DefStore.defs.get("/someURL")).to.be("someData");
	});

	it("should handle DefsFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefsFetched("r", "v", "q", ["someData"]));
		expect(DefStore.defs.list("r", "v", "q")).to.eql(["someData"]);
	});

	it("should handle HighlightDef", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("someDef"));
		expect(DefStore.highlightedDef).to.be("someDef");

		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef(null));
		expect(DefStore.highlightedDef).to.be(null);
	});

	it("should handle RefsFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.RefsFetched("/someURL", "f", ["someData"]));
		expect(DefStore.refs.get("/someURL", "f")).to.eql(["someData"]);
	});
});
