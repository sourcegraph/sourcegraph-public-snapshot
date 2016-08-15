import expect from "expect.js";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoStore} from "sourcegraph/repo/RepoStore";

describe("RepoStore", () => {
	it("should handle FetchedCommit", () => {
		RepoStore.directDispatch(new RepoActions.FetchedCommit("r", "v", {ID: "c"}));
		expect(RepoStore.commits.get("r", "v")).to.eql({ID: "c"});
	});
});
