import expect from "expect.js";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeStore} from "sourcegraph/tree/TreeStore";

describe("TreeStore", () => {
	it("should handle CommitFetched", () => {
		TreeStore.directDispatch(new TreeActions.CommitFetched("aRepo", "aRev", "aPath", "someResults"));
		expect(TreeStore.commits.get("aRepo", "aRev", "aPath")).to.be("someResults");
	});
});
