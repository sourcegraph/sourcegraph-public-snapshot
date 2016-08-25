import expect from "expect.js";
import * as React from "react";
import {BlobMain} from "sourcegraph/blob/BlobMain";
import {renderToString} from "sourcegraph/util/componentTestUtils";

describe("BlobMain", () => {
	it("should show an error page if the blob failed to load", () => {
		let o = renderToString(<BlobMain location={{key: "", pathname: "", search: "", action: "", query: {}, state: {}}}
			repo="r" blob={{Error: {response: {status: 500}}}} />, {router: {}});
		expect(o).to.contain("is not available");
	});
});
