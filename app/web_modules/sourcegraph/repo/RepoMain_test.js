// @flow weak

import React from "react";
import expect from "expect.js";
import RepoMain from "sourcegraph/repo/RepoMain";
import {renderToString} from "sourcegraph/util/componentTestUtils";
import {render} from "sourcegraph/util/renderTestUtils";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as BuildActions from "sourcegraph/build/BuildActions";

describe("RepoMain", () => {
	it("should show an error page if the repo failed to load", () => {
		let o = renderToString(<RepoMain repo="r" repoObj={{Error: true}} />);
		expect(o).to.contain("is not available");
	});

	it("should show an error page if the rev failed to resolve/load", () => {
		const o = renderToString(<RepoMain repo="r" rev="v" resolvedRev={{Error: {}}} />);
		expect(o).to.contain("Revision is not available");
	});

	describe("repo creation", () => {
		it("should trigger WantCreateRepo for just-resolved remote repos", () => {
			const res = render(<RepoMain repo="r" repoResolution={{RemoteRepo: {GitHubID: 123}}} location={{pathname: "/r", state: {}}} />);
			expect(res.actions).to.eql([new RepoActions.WantCreateRepo("r", {GitHubID: 123})]);
		});
	});

	describe("build creation", () => {
		it("should trigger CreateBuild when there is no build", () => {
			const res = render(<RepoMain repo="r" commitID="c" rev="v" build={{Error: {response: {status: 404}}}} location={{pathname: "/r", state: {}}} />);
			expect(res.actions).to.eql([new BuildActions.CreateBuild("r", "c", "v", null)]);
		});
	});
});
