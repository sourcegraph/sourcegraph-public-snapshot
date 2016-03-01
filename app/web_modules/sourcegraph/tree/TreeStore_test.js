import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

describe("TreeStore", () => {
	it("should handle ResultsFetched", () => {
		Dispatcher.directDispatch(TreeStore, new TreeActions.CommitFetched("aRepo", "aRev", "aPath", "someResults"));
		expect(TreeStore.commits.get("aRepo", "aRev", "aPath")).to.be("someResults");
	});
});
