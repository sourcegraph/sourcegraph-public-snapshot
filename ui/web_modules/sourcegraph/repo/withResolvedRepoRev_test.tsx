// tslint:disable: typedef ordered-imports

import * as React from "react";
import expect from "expect.js";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {render} from "sourcegraph/util/renderTestUtils";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import * as RepoActions from "sourcegraph/repo/RepoActions";

const C = withResolvedRepoRev((props) => null, true);

describe("withResolvedRepoRev", () => {
	it("should render initially", () => {
		render(<C params={{splat: "r"}} location={{} as HistoryModule.Location} />);
	});

	it("should render if the repo and rev exist", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
		render(<C params={{splat: "r"}}  location={{} as HistoryModule.Location}/>);
	});

	it("should render if the repo is cloning", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
		RepoStore.directDispatch(new RepoActions.RepoCloning("r", true));
		render(<C params={{splat: "r"}}  location={{} as HistoryModule.Location}/>);
	});

	it("should render if the repo does not exist", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {Error: true}));
		render(<C params={{splat: "r"}}  location={{} as HistoryModule.Location}/>);
	});

	describe("repo resolution", () => {
		it("should initially trigger WantResolveRepo", () => {
			const res = render(<C params={{splat: "r"}}  location={{} as HistoryModule.Location}/>, {router: {}});
			expect(res.actions).to.eql([new RepoActions.WantResolveRepo("r")]);
		});
		it("should trigger WantRepo for resolved local repos", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("r", {Repo: 1, CanonicalPath: "r"}));
			let calledReplace = false;
			const res = render(<C params={{splat: "r"}}  location={{} as HistoryModule.Location}/>, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(false);
			expect(res.actions).to.eql([new RepoActions.WantRepo("r")]);
		});
		it("should NOT trigger WantRepo for resolved remote repos", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("github.com/user/repo", {RemoteRepo: {Owner: "user", Name: "repo"}}));
			let calledReplace = false;
			const res = render(<C params={{splat: "github.com/user/repo"}}  location={{} as HistoryModule.Location}/>, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(false);
			expect(res.actions).to.eql([]);
		});

		it("should redirect for resolved local repos with different canonical name", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("repo", {Repo: 1, CanonicalPath: "renamedRepo"}));
			let calledReplace = false;
			render(<C params={{splat: "repo"}} location={{pathname: "sg.com/alias"} as HistoryModule.Location} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(true);
		});
		it("should redirect for resolved remote repos with different canonical name", () => {
			RepoStore.directDispatch(new RepoActions.RepoResolved("github.com/user/repo", {RemoteRepo: {Owner: "renamedUser", Name: "renamedRepo"}}));
			let calledReplace = false;
			render(<C params={{splat: "github.com/user/repo"}} location={{pathname: "sg.com/alias"} as HistoryModule.Location} />, {
				router: {replace: () => calledReplace = true},
			});
			expect(calledReplace).to.be(true);
		});
	});
});
