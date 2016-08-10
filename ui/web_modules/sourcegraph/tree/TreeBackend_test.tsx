// tslint:disable: typedef ordered-imports

import expect from "expect.js";

import * as Dispatcher from "sourcegraph/Dispatcher";
import {TreeBackend} from "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {immediateSyncPromise} from "sourcegraph/util/immediateSyncPromise";

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
		})).to.eql([new TreeActions.CommitFetched(entry.repo, entry.rev, entry.path, "someTreeCommit")]);
	});

	it("should handle WantFileList", () => {
		const repo = "aRepo";
		const commitID = "aCommitID";
		let expectedURI = `/.api/repos/${repo}@${commitID}/-/tree-list`;

		TreeBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => ({Files: ["a", "b"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantFileList(repo, commitID));
		})).to.eql([
			new RepoActions.RepoCloning(repo, false),
			new TreeActions.FileListFetched(repo, commitID, {Files: ["a", "b"]}),
		]);
	});
});
