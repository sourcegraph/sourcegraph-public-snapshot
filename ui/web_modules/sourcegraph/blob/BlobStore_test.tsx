// tslint:disable

import expect from "expect.js";

import {BlobStore, keyForFile} from "sourcegraph/blob/BlobStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";

describe("BlobStore", () => {
	it("should handle FileFetched", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("aRepo", "aRev", "aPath", "someContent" as any));
		expect(BlobStore.files[keyForFile("aRepo", "aRev", "aPath")]).to.be("someContent");
	});
});
