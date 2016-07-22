import expect from "expect.js";

import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", "someData"));
		expect(DefStore.defs.get("r", "v", "d")).to.be("someData");
	});

	it("should handle DefsFetched", () => {
		DefStore.directDispatch(new DefActions.DefsFetched("r", "v", "q", "fpp", ["someData"]));
		expect(DefStore.defs.list("r", "v", "q", "fpp")).to.eql(["someData"]);
	});

	it("should handle Hovering", () => {
		let pos = {repo: "foo", commit: "aaa", file: "bar", line: 42, character: 3};
		DefStore.directDispatch(new DefActions.Hovering(pos));
		expect(DefStore.hoverPos).to.be(pos);

		DefStore.directDispatch(new DefActions.Hovering(null));
		expect(DefStore.hoverPos).to.be(null);
	});

	it("should handle RefLocationsFetched", () => {
		DefStore.directDispatch(new DefActions.RefLocationsFetched(
			new DefActions.WantRefLocations({repo: "r", rev: "v", def: "d", repos: []}), ["someData"],
		));
		expect(DefStore.getRefLocations({repo: "r", rev: "v", def: "d", repos: []})).to.be.ok();
	});

	it("should handle RefsFetched", () => {
		DefStore.directDispatch(new DefActions.RefsFetched("r", "v", "d", "rr", "rf", ["someData"]));
		expect(DefStore.refs.get("r", "v", "d", "rr", "rf")).to.eql(["someData"]);
	});
});
