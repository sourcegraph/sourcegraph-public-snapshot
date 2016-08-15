import expect from "expect.js";
import * as DefActions from "sourcegraph/def/DefActions";
import {DefStore} from "sourcegraph/def/DefStore";
import {Def} from "sourcegraph/def/index";

describe("DefStore", () => {
	it("should handle DefFetched", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {Path: "somePath"} as Def));
		expect(DefStore.defs.get("r", "v", "d")).to.eql({Path: "somePath"});
	});

	it("should handle DefsFetched", () => {
		DefStore.directDispatch(new DefActions.DefsFetched("r", "v", "q", "fpp", [{Path: "somePath"} as Def]));
		expect(DefStore.defs.list("r", "v", "q", "fpp")).to.eql([{Path: "somePath"}]);
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
			new DefActions.WantRefLocations({repo: "r", rev: "v", commitID: "c", def: "d", repos: []}), ["someData"],
		));
		expect(DefStore.getRefLocations({repo: "r", rev: "v", commitID: "c", def: "d", repos: []})).to.be.ok();
	});

	it("should handle RefsFetched", () => {
		DefStore.directDispatch(new DefActions.RefsFetched("r", "v", "d", "rr", "rf", ["someData"]));
		expect(DefStore.refs.get("r", "v", "d", "rr", "rf")).to.eql(["someData"]);
	});
});
