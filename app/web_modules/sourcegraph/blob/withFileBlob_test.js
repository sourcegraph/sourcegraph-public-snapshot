// @flow weak

import React from "react";
import expect from "expect.js";
import withFileBlob from "sourcegraph/blob/withFileBlob";
import {renderedStatus} from "sourcegraph/app/statusTestUtils";
import BlobStore from "sourcegraph/blob/BlobStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {rel as relPath} from "sourcegraph/app/routePatterns";
import {render} from "sourcegraph/util/renderTestUtils";

const C = withFileBlob((props) => null);

const props = {
	params: {splat: [null, "f"]},
	route: {path: relPath.blob},
};

describe("withFileBlob", () => {
	describe("status", () => {
		it("should have no error initially", () => {
			expect(renderedStatus(
				<C repo="r" rev="v" {...props} />
			)).to.eql({error: null});
		});

		it("should have no error if the blob and rev exist", () => {
			BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {CommitID: "c"}));
			expect(renderedStatus(
				<C repo="r" rev="v" commitID="c" {...props} />
			)).to.eql({error: null});
		});

		it("should have error if the blob does not exist", () => {
			BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {Error: true}));
			expect(renderedStatus(
				<C repo="r" rev="v" {...props} />
			)).to.eql({error: true});
		});
	});
	it("should redirect to the tree URL when the blob is a tree", (done) => {
		BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {Entries: [], CommitID: "c"}));
		let calledReplace = false;
		render(<C repo="r" rev="v" path="f" {...props} />, {
			router: {replace: () => calledReplace = true},
		});
		setTimeout(() => {
			expect(calledReplace).to.be(true);
			done();
		});
	});
});
