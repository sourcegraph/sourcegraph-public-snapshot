import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

afterEach(GitHubReposStore.reset.bind(GitHubReposStore));
beforeEach(GitHubReposStore.reset.bind(GitHubReposStore));

describe("GitHubReposStore", () => {
	it("should handle Mirror Repo Added", () => {
		Dispatcher.directDispatch(GitHubReposStore, new DashboardActions.MirrorRepoAdded( "repo", {RemoteRepos: "hello"}));
		expect(GitHubReposStore.remoteRepos.repos).to.be("hello");
	});
});
