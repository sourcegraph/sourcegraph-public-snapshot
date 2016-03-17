import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import TreeBackend from "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";

describe("TreeBackend", () => {
	it("should handle WantCommit", () => {
		let entry = {
			repo: "aRepo",
			rev: "aRev",
			path: "aPath",
		};
		let expectedURI = `/.api/repos/${entry.repo}/.commits?Head=${entry.rev}&Path=${entry.path}&PerPage=1`;

		TreeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, {Commits: ["someTreeCommit"]});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantCommit(entry.repo, entry.rev, entry.path));
		})).to.eql([new TreeActions.CommitFetched(entry.repo, entry.rev, entry.path, {Commits: ["someTreeCommit"]})]);
	});

	it("should handle WantFileList", () => {
		const repo = "aRepo";
		const rev = "aRev";
		const commitID = "aCommit";
		let expectedURI = `/.api/repos/${repo}@${rev}===${commitID}/.tree-list`;

		TreeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, {Files: ["a", "b"]});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantFileList(repo, rev, commitID));
		})).to.eql([new TreeActions.FileListFetched(repo, rev, commitID, {Files: ["a", "b"]})]);
	});
});
