// @flow weak

import * as React from "react";
import expect from "expect.js";
import RepoMain from "sourcegraph/repo/RepoMain";
import {renderToString} from "sourcegraph/util/componentTestUtils";
import {render} from "sourcegraph/util/renderTestUtils";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {Origin_GitHub} from "sourcegraph/repo/RepoBackend";

const common = {
	routes: [],
	route: {},
};

describe("RepoMain", () => {
	it("should show an error page if the repo failed to load", () => {
		let o = renderToString(<RepoMain repo="r" repoObj={{Error: true}} {...common} />);
		expect(o).to.contain("is not available");
	});

	it("should show an error page if the rev failed to resolve/load", () => {
		const o = renderToString(<RepoMain repo="r" rev="v" resolvedRev={{Error: {}}} {...common} />);
		expect(o).to.contain("Revision is not available");
	});

	describe("repo creation", () => {
		it("should trigger WantCreateRepo for just-resolved remote repos", () => {
			const res = render(<RepoMain repo="r" repoResolution={{RemoteRepo: {Origin: {ID: "123", Service: Origin_GitHub}}}} location={{pathname: "/r", state: {}}} {...common} />);
			expect(res.actions).to.eql([new RepoActions.WantCreateRepo("r", {Origin: {ID: "123", Service: Origin_GitHub}})]);
		});
	});

	describe("build creation", () => {
		it("should trigger CreateBuild when there is no build", () => {
			const res = render(<RepoMain repo="r" commitID="c" rev="v" build={{Error: {response: {status: 404}}}} location={{pathname: "/r", state: {}}} {...common} />);
			expect(res.actions).to.eql([new BuildActions.CreateBuild("r", "c", "v", null)]);
		});
	});
});
