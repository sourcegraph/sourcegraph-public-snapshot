// @flow weak

import React from "react";
import expect from "expect.js";
import withResolvedRepoRev from "sourcegraph/repo/withResolvedRepoRev";
import {renderedStatus} from "sourcegraph/app/statusTestUtils";
import {render} from "sourcegraph/util/renderTestUtils";
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

		it("should be HTTP 202 if the repo is cloning", () => {
			RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
			RepoStore.directDispatch(new RepoActions.RepoCloning("r", true));
			expect(renderedStatus(
				<C params={{splat: "r"}} />
			)).to.eql({error: {status: 202}});
		});

		it("should have error if the repo does not exist", () => {
			RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {Error: true}));
			expect(renderedStatus(
				<C params={{splat: "r"}} />
			)).to.eql({error: true});
		});
	});

	describe("repo resolution", () => {
		it("should NOT initially trigger WantResolveRepo (the route onEnter/onChange does it)", () => {
			const res = render(<C params={{splat: "r"}} />, {router: {}});
			expect(res.actions).to.eql([]);
		});
		it("should trigger WantRepo for resolved local repos", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("r", {Result: {Repo: {URI: "r"}}}));
			let calledReplace = false;
			const res = render(<C params={{splat: "r"}} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(false);
			expect(res.actions).to.eql([new RepoActions.WantRepo("r")]);
		});
		it("should NOT trigger WantRepo for resolved remote repos", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("github.com/user/repo", {Result: {RemoteRepo: {Owner: "user", Name: "repo"}}}));
			let calledReplace = false;
			const res = render(<C params={{splat: "github.com/user/repo"}} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(false);
			expect(res.actions).to.eql([]);
		});

		it("should redirect for resolved local repos with different canonical name", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("repo", {Result: {Repo: {URI: "renamedRepo"}}}));
			let calledReplace = false;
			render(<C params={{splat: "repo"}} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(true);
		});
		it("should redirect for resolved remote repos with different canonical name", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("github.com/user/repo", {Result: {RemoteRepo: {Owner: "renamedUser", Name: "renamedRepo"}}}));
			let calledReplace = false;
			render(<C params={{splat: "github.com/user/repo"}} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(true);
		});
	});
});
