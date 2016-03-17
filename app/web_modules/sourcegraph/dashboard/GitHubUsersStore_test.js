import expect from "expect.js";

import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

afterEach(GitHubUsersStore.reset.bind(GitHubUsersStore));
beforeEach(GitHubUsersStore.reset.bind(GitHubUsersStore));

describe("GitHubUsersStore", () => {
	it("should handle user invites", () => {
		GitHubUsersStore.directDispatch(new DashboardActions.UsersInvited({Users: "hello"}));
		expect(GitHubUsersStore.users.users).to.be("hello");
	});
});
