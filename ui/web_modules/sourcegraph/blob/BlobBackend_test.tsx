import expect from "expect.js";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobBackend} from "sourcegraph/blob/BlobBackend";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {immediateSyncPromise} from "sourcegraph/util/testutil/immediateSyncPromise";

describe("BlobBackend", () => {
	it("should handle WantFile", () => {
		BlobBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/repos/aRepo@aCommitID/-/tree/aPath?ContentsAsString=true&NoSrclibAnns=true");
			return immediateSyncPromise({
				status: 200,
				json: () => "someFile",
			});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "aCommitID", "aPath"));
		})).to.eql([
			new RepoActions.RepoCloning("aRepo", false),
			new BlobActions.FileFetched("aRepo", "aCommitID", "aPath", "someFile" as any),
		]);
	});
});
