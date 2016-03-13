import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import TreeBackend from "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";

describe("TreeBackend", () => {
	it("should handle WantCommit", () => {
		let entry = {
			repo: "aRepo",
			rev: "aRev",
			path: "aPath",
		};
		let expectedURI = `/.api/repos/${entry.repo}/-/commits?Head=${entry.rev}&Path=${entry.path}&PerPage=1`;

		TreeBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => ({Commits: ["someTreeCommit"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantCommit(entry.repo, entry.rev, entry.path));
		})).to.eql([new TreeActions.CommitFetched(entry.repo, entry.rev, entry.path, {Commits: ["someTreeCommit"]})]);
	});

	it("should handle WantFileList", () => {
		const repo = "aRepo";
		const rev = "aRev";
		let expectedURI = `/.api/repos/${repo}@${rev}/-/tree-list`;

		TreeBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => ({Files: ["a", "b"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantFileList(repo, rev));
		})).to.eql([new TreeActions.FileListFetched(repo, rev, {Files: ["a", "b"]})]);
	});
});
