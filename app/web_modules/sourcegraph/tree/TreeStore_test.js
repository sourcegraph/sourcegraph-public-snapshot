import expect from "expect.js";

import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

describe("TreeStore", () => {
	it("should handle CommitFetched", () => {
		TreeStore.directDispatch(new TreeActions.CommitFetched("aRepo", "aRev", "aPath", "someResults"));
		expect(TreeStore.commits.get("aRepo", "aRev", "aPath")).to.be("someResults");
	});

	it("should handle FileListFetched", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("aRepo", "aRev", {Files: ["someResults"]}));
		expect(TreeStore.fileLists.get("aRepo", "aRev")).to.have.property("Files");
	});

	it("should not crash on special directory names", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("aRepo", "aRev", {Files: ["constructor/file.txt"]}));
	});
});
