// tslint:disable: typedef ordered-imports curly

import * as React from "react";
import expect from "expect.js";
import {withFileBlob} from "sourcegraph/blob/withFileBlob";
import {render} from "sourcegraph/util/renderTestUtils";
import {BlobStore} from "sourcegraph/blob/BlobStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import {rel as relPath} from "sourcegraph/app/routePatterns";

const C = withFileBlob((props) => null);

const props = {
	params: {splat: [null, "f"]},
	route: {path: relPath.blob},
};

describe("withFileBlob", () => {
	it("should render initially", () => {
		render(<C repo="r" rev="v" {...props} />);
	});

	it("should render when the blob and commitID exist", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {CommitID: "c"} as any));
		render(<C repo="r" rev="v" commitID="c" {...props} />);
	});

	it("should render when the blob does not exist", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {Error: true} as any));
		render(<C repo="r" rev="v" commitID="c" {...props} />);
	});
	it("should redirect to the tree URL when the blob is a tree", (done) => {
		BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "f", {Entries: [], CommitID: "c"} as any));
		let calledReplace = false;
		render(<C repo="r" commitID="v" path="f" {...props} />, {
			router: {replace: () => calledReplace = true},
		});
		setTimeout(() => {
			expect(calledReplace).to.be(true);
			done();
		});
	});
});
