	import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

describe("DashboardStore", () => {
	it("should handle Repo Created", () => {
		Dispatcher.directDispatch(DashboardStore, new DashboardActions.RepoCreated("repos"));
		expect(DashboardStore.repos).to.eql("repos");
	});

	it("should handle User Invited", () => {
		Dispatcher.directDispatch(DashboardStore, new DashboardActions.UserInvited("aUser"));
		expect(DashboardStore.users).to.eql(["aUser"]);
	});
});
