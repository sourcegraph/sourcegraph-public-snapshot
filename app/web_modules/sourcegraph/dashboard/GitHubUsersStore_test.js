import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

afterEach(GitHubUsersStore.reset.bind(GitHubUsersStore));
beforeEach(GitHubUsersStore.reset.bind(GitHubUsersStore));

describe("DefStore", () => {
	it("should handle Mirror Repo Added", () => {
		Dispatcher.directDispatch(GitHubUsersStore, new DashboardActions.UsersInvited({Users: "hello"}));
		expect(GitHubUsersStore.users.users).to.be("hello");
	});
});
