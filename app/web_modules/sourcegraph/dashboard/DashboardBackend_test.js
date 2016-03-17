import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DashboardBackend from "sourcegraph/dashboard/DashboardBackend";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

describe("DashboardBackend", () => {
	it("should handle WantInviteUsers", () => {
		let expectedURI = "/.ui/.invite-bulk";

		DashboardBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			expect(options.json).to.have.property("Emails");
			callback(null, {statusCode: 200}, "someFile");
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			Dispatcher.directDispatch(DashboardBackend, new DashboardActions.WantInviteUsers("someEmails"));
		})).to.eql([new DashboardActions.UsersInvited("someFile")]);
	});
});
