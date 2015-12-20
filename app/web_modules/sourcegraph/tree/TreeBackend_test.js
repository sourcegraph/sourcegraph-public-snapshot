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
		let expectedURI = `/.ui/${entry.repo}/.commits?Head=${entry.rev}&Path=${entry.path}&PerPage=1`;

		TreeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, {Commits: ["someTreeCommit"]});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(TreeBackend, new TreeActions.WantCommit(entry.repo, entry.rev, entry.path));
		})).to.eql([new TreeActions.CommitFetched(entry.repo, entry.rev, entry.path, {Commits: ["someTreeCommit"]})]);
	});
});
