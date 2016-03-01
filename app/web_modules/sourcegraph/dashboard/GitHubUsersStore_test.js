import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

describe("DefStore", () => {
	it("should handle Want Add Mirror Repos", () => {
		Dispatcher.directDispatch(GitHubUsersStore, new DashboardActions.WantInviteUsers());
		expect(GitHubUsersStore.showLoading).to.be(true);
	});

	it("should handle Mirror Repo Added", () => {
		Dispatcher.directDispatch(GitHubUsersStore, new DashboardActions.UsersInvited({Users: "hello"}));
		expect(GitHubUsersStore.users.users).to.be("hello");
		expect(GitHubUsersStore.showLoading).to.be(false);
	});
});
