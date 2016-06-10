// @flow

import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DeltaBackend from "sourcegraph/delta/DeltaBackend";
import * as DeltaActions from "sourcegraph/delta/DeltaActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";

const d = {Base: {Repo: 1, CommitID: "cbase"}, Head: {Repo: 1, CommitID: "chead"}};

describe("DeltaBackend", () => {
	it("should handle WantFiles", () => {
		DeltaBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos/1@chead/-/delta/cbase/-/files");
			return immediateSyncPromise({status: 200, json: () => ({Ds: d})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DeltaBackend.__onDispatch(new DeltaActions.WantFiles(d.Base.Repo, d.Base.CommitID, d.Head.Repo, d.Head.CommitID));
		})).to.eql([new DeltaActions.FetchedFiles(d.Base.Repo, d.Base.CommitID, d.Head.Repo, d.Head.CommitID, {Ds: d})]);
	});
});
