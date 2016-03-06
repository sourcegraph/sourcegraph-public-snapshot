import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

afterEach(DashboardStore.reset.bind(DashboardStore));
beforeEach(DashboardStore.reset.bind(DashboardStore));

describe("DashboardStore", () => {
	it("should handle User Invited", () => {
		Dispatcher.directDispatch(DashboardStore, new DashboardActions.UserInvited("aUser"));
		expect(DashboardStore.users).to.eql(["aUser"]);
	});
});
