import expect from "expect.js";
import * as React from "react";
import {RepoMain} from "sourcegraph/repo/RepoMain";
import {renderToString} from "sourcegraph/util/testutil/componentTestUtils";

const common = {
	routes: [],
	route: {},
};

describe("RepoMain", () => {
	it("should show an error page if the repo failed to load", () => {
		let o = renderToString(<RepoMain repo="r" repoObj={{Error: true}} commit={{} as GQL.ICommitState} {...common} relay={null} />);
		expect(o).to.contain("is not available");
	});

	it("should show an error page if the rev failed to resolve/load", () => {
		const o = renderToString(<RepoMain repo="r" rev="v" commit={{} as GQL.ICommitState} {...common} relay={null} />);
		expect(o).to.contain("Revision is not available");
	});
});
