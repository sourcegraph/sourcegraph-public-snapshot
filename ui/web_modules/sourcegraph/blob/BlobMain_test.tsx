// tslint:disable: typedef ordered-imports

import * as React from "react";
import expect from "expect.js";
import {BlobMain} from "sourcegraph/blob/BlobMain";
import {renderToString} from "sourcegraph/util/componentTestUtils";

describe("BlobMain", () => {
	it("should show an error page if the blob failed to load", () => {
		let o = renderToString(<BlobMain repo="r" blob={{Error: {response: {status: 500}}}} />, {router: {}});
		expect(o).to.contain("is not available");
	});
});
