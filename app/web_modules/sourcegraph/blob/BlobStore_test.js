import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import BlobStore from "sourcegraph/blob/BlobStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";

describe("BlobStore", () => {
	it("should handle FileFetched", () => {
		Dispatcher.directDispatch(BlobStore, new BlobActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(BlobStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
