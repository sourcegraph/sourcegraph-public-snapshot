import expect from "expect.js";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeBackend} from "sourcegraph/tree/TreeBackend";
import {immediateSyncPromise} from "sourcegraph/util/testutil/immediateSyncPromise";

describe("TreeBackend", () => {
	it("should handle WantCommit", () => {
		let entry = {
			repo: "aRepo",
			rev: "aRev",
			path: "aPath",
		};
		let expectedURI = `/.api/repos/${entry.repo}/-/commits?Head=${entry.rev}&Path=${entry.path}&PerPage=1`;

		TreeBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => ({Commits: ["someTreeCommit"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			TreeBackend.__onDispatch(new TreeActions.WantCommit(entry.repo, entry.rev, entry.path));
		})).to.eql([new TreeActions.CommitFetched(entry.repo, entry.rev, entry.path, "someTreeCommit")]);
	});
});
