// @flow weak

import React from "react";
import expect from "expect.js";
import withResolvedRepoRev from "sourcegraph/repo/withResolvedRepoRev";
import {renderedStatus} from "sourcegraph/app/statusTestUtils";
import RepoStore from "sourcegraph/repo/RepoStore";
import * as RepoActions from "sourcegraph/repo/RepoActions";

const C = withResolvedRepoRev((props) => null);

describe("withResolvedRepoRev", () => {
	describe("status", () => {
		it("should have no error initially", () => {
			expect(renderedStatus(
				<C params={{splat: "r"}} />
			)).to.eql({error: null});
		});

		it("should have no error if the repo and rev exist", () => {
			RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
			expect(renderedStatus(
				<C params={{splat: "r"}} />
			)).to.eql({error: null});
		});

		it("should have error if the repo does not exist", () => {
			RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {Error: true}));
			expect(renderedStatus(
				<C params={{splat: "r"}} />
			)).to.eql({error: true});
		});
	});
});
