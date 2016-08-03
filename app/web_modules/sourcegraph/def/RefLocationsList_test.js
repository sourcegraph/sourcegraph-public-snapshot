import autotest from "sourcegraph/util/autotest";

import * as React from "react";
import expect from "expect.js";
import RefLocationsList from "sourcegraph/def/RefLocationsList";
import {renderToString} from "sourcegraph/util/componentTestUtils";
import testdataData from "sourcegraph/def/testdata/RefLocationsList-data.json";
import testdataEmpty from "sourcegraph/def/testdata/RefLocationsList-empty.json";

const ctx = {
	eventLogger: {},
	githubToken: {},
	signedIn: false,
};

describe("RefLocationsList", () => {
	it("should render definition data", () => {
		autotest(testdataData, "sourcegraph/def/testdata/RefLocationsList-data.json",
			<RefLocationsList
				repo="r" rev="v" path="p"
				def={{Repo: "r", CommitID: "c"}}
				location={{}}
				refLocations={{RepoRefs: [{Repo: "r", Files: [{Path: "f", Count: 2}]}]}} />,
			Object.assign({}, ctx, {signedIn: true, githubToken: {scope: "repo"}}),
		);
	});

	it("should render empty", () => {
		autotest(testdataEmpty, "sourcegraph/def/testdata/RefLocationsList-empty.json",
			<RefLocationsList
				repo="r" rev="v" path="p"
				def={{Repo: "r", CommitID: "c"}}
				location={{}}
				refLocations={{}} />,
			ctx,
		);
	});

	describe("signup and private repo CTA", () => {
		const refLocsMoreThan1 = {
			RepoRefs: [
				{Repo: "r", Files: [{Path: "f", Count: 2}]},
				{Repo: "r2", Files: [{Path: "f2", Count: 4}]},
			],
		};

		it("should show signup CTA if not authed", () => {
			const s = renderToString(
				<RefLocationsList repo="r" rev="v" def={{Repo: "r", CommitID: "c"}}	location={{}} refLocations={refLocsMoreThan1} />,
				Object.assign({}, ctx, {signedIn: false}),
			);
			expect(s).to.contain("Sign in");
			expect(s).to.not.contain("Authorize");
		});

		it("should show private repo CTA if signed up but not private-repo authed", () => {
			const s = renderToString(
				<RefLocationsList repo="r" rev="v" def={{Repo: "r", CommitID: "c"}}	location={{}} refLocations={refLocsMoreThan1} />,
				Object.assign({}, ctx, {signedIn: true, githubToken: {scope: ""}}), // no "repo" scope
			);
			expect(s).to.contain("Authorize");
			expect(s).to.not.contain("Sign in");
		});

		it("should NOT show CTA if already authed", () => {
			const s = renderToString(
				<RefLocationsList repo="r" rev="v" def={{Repo: "r", CommitID: "c"}}	location={{}} refLocations={refLocsMoreThan1} />,
				Object.assign({}, ctx, {signedIn: true, githubToken: {scope: "repo"}}), // private repo scope
			);
			expect(s).to.not.contain("Authorize");
			expect(s).to.not.contain("Sign in");
		});
	});
});
