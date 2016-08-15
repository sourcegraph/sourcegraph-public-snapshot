import expect from "expect.js";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobStore, keyForFile} from "sourcegraph/blob/BlobStore";

describe("BlobStore", () => {
	it("should handle FileFetched", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("aRepo", "aRev", "aPath", "someContent" as any));
		expect(BlobStore.files[keyForFile("aRepo", "aRev", "aPath")]).to.be("someContent");
	});
});
