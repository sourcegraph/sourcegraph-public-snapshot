import expect from "expect.js";

import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

describe("TreeStore", () => {
	it("should handle ResultsFetched", () => {
		TreeStore.directDispatch(new TreeActions.CommitFetched("aRepo", "aRev", "aPath", "someResults"));
		expect(TreeStore.commits.get("aRepo", "aRev", "aPath")).to.be("someResults");
	});

	it("should handle ResultsFetched", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("aRepo", "aRev", "aCommit", {Files: ["someResults"]}));
		expect(TreeStore.fileLists.get("aRepo", "aRev", "aCommit")).to.have.property("Files");
	});
});
