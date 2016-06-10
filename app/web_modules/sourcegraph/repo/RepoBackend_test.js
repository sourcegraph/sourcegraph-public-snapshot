import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import RepoBackend from "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";

describe("RepoBackend", () => {
	it("should handle WantCommit", () => {
		RepoBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos/r@v/-/commit");
			return immediateSyncPromise({status: 200, json: () => ({ID: "c"})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			RepoBackend.__onDispatch(new RepoActions.WantCommit("r", "v"));
		})).to.eql([new RepoActions.FetchedCommit("r", "v", {ID: "c"})]);
	});
});
