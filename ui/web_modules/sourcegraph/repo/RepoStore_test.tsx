// tslint:disable: typedef ordered-imports

import expect from "expect.js";

import {RepoStore} from "sourcegraph/repo/RepoStore";
import * as RepoActions from "sourcegraph/repo/RepoActions";

describe("RepoStore", () => {
	it("should handle FetchedCommit", () => {
		RepoStore.directDispatch(new RepoActions.FetchedCommit("r", "v", {ID: "c"}));
		expect(RepoStore.commits.get("r", "v")).to.eql({ID: "c"});
	});
});
