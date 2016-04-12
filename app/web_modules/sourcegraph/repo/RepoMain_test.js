// @flow weak

import React from "react";
import expect from "expect.js";
import RepoMain from "sourcegraph/repo/RepoMain";
import {renderToString} from "sourcegraph/util/componentTestUtils";
import {render} from "sourcegraph/util/renderTestUtils";
import * as RepoActions from "sourcegraph/repo/RepoActions";

describe("RepoMain", () => {
	it("should show an error page if the repo failed to load", () => {
		let o = renderToString(<RepoMain repo="r" repoObj={{Error: true}} />);
		expect(o).to.contain("is not available");
	});

	describe("repo creation", () => {
		it("should trigger WantCreateRepo for just-resolved remote repos", () => {
			const res = render(<RepoMain repo="r" repoResolution={{Result: {RemoteRepo: {GitHubID: 123}}}} />);
			expect(res.actions).to.eql([new RepoActions.WantCreateRepo("r", {Op: {FromGitHubID: 123}})]);
		});
	});
});
