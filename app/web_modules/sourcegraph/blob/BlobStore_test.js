import expect from "expect.js";

import BlobStore from "sourcegraph/blob/BlobStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";

describe("BlobStore", () => {
	it("should handle FileFetched", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(BlobStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
