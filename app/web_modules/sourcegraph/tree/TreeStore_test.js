import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

afterEach(TreeStore.reset.bind(TreeStore));
beforeEach(TreeStore.reset.bind(TreeStore));

describe("TreeStore", () => {
	it("should handle ResultsFetched", () => {
		Dispatcher.directDispatch(TreeStore, new TreeActions.CommitFetched("aRepo", "aRev", "aPath", "someResults"));
		expect(TreeStore.commits.get("aRepo", "aRev", "aPath")).to.be("someResults");
	});
});
