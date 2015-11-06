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

	it("should handle ExampleFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.ExampleFetched("/someURL", 42, "someData"));
		expect(DefStore.examples.get("/someURL", 42)).to.be("someData");
	});

	it("should handle NoExampleAvailable", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.NoExampleAvailable("/someURL", 50));
		Dispatcher.directDispatch(DefStore, new DefActions.NoExampleAvailable("/someURL", 42));
		Dispatcher.directDispatch(DefStore, new DefActions.NoExampleAvailable("/someURL", 100));
		expect(DefStore.examples.getCount("/someURL")).to.be(42);
	});

	it("should handle DiscussionsFetched", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DiscussionsFetched("/someURL", "someData"));
		expect(DefStore.discussions.get("/someURL")).to.be("someData");
	});
});
