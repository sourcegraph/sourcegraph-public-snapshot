import expect from "expect.js";
import * as React from "react";
import {RepoMain} from "sourcegraph/repo/RepoMain";
import {renderToString} from "sourcegraph/util/testutil/componentTestUtils";

const common = {
	routes: [],
	route: {},
	location: {},
};

describe("RepoMain", () => {
	it("should show an error page if the repo failed to load", () => {
		let o = renderToString(<RepoMain repo="r" rev="v" repository={null} commit={{} as GQL.ICommitState} {...common} relay={null} params={undefined as any} />);
		expect(o).to.contain("Repository not found.");
	});

	it("should show an error page if the rev failed to resolve/load", () => {
		let o = renderToString(<RepoMain repo="r" rev="v" repository={{} as GQL.IRepository} commit={{} as GQL.ICommitState} {...common} relay={null} params={undefined as any} />);
		expect(o).to.contain("Revision is not available");
	});
});
