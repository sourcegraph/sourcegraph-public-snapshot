import expect from "expect.js";
import {Commit} from "sourcegraph/api";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoStore} from "sourcegraph/repo/RepoStore";

describe("RepoStore", () => {
	it("should handle FetchedCommit", () => {
		RepoStore.directDispatch(new RepoActions.FetchedCommit("r", "v", {ID: "c"} as any as Commit));
		expect(RepoStore.commits.get("r", "v")).to.eql({ID: "c"});
	});
});
