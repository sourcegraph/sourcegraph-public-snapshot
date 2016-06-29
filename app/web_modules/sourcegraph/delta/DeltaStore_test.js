// @flow

import expect from "expect.js";

import DeltaStore from "sourcegraph/delta/DeltaStore";
import * as DeltaActions from "sourcegraph/delta/DeltaActions";

const d = {Base: {Repo: 1, CommitID: "cbase"}, Head: {Repo: 2, CommitID: "chead"}};

describe("DeltaStore", () => {
	it("should handle FetchedCommit", () => {
		DeltaStore.directDispatch(new DeltaActions.FetchedFiles(d.Base.Repo, d.Base.CommitID, d.Head.Repo, d.Head.CommitID, {Ds: d}));
		expect(DeltaStore.files.get(d.Base.Repo, d.Base.CommitID, d.Head.Repo, d.Head.CommitID)).to.eql({Ds: d});
	});
});
