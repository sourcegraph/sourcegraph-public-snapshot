import autotest from "sourcegraph/util/autotest";

import React from "react";
import moment from "moment";
import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";

import testdataData from "sourcegraph/dashboard/testdata/DashboardUsers-data.json";
import testdataInvited from "sourcegraph/dashboard/testdata/DashboardUsers-InvitedUsers.json";
import testdataNotInvited from "sourcegraph/dashboard/testdata/DashboardUsers-notInvited.json";

describe("DashboardUsers", () => {
	it("should render a User", () => {
		let remoteAccount = {Name: "RemoteName", AvatarURL: "asdf"};
		let localAccount = {Name: "Local Loco", AvatarURL: "zxcv", UID: "1"};
		autotest(testdataData, `${__dirname}/testdata/DashboardUsers-data.json`,
			<DashboardUsers users={[{Email:"abc", IsInvited: true, RemoteAccount: remoteAccount, LocalAccount: localAccount}]}
				currentUser={{UID: "1", Admin: false}}
				onboarding={{LinkGitHub: false}}
				allowStandaloneUsers={false}
				allowGitHubUsers={false} />
		);
	});
});

describe("DashboardUsers", () => {
	it("should render invited users", () => {
		let invitedUser1 = {Email: "abc1", IsInvited: true, RemoteAccount: {Name: "xyz1"}};
		let invitedUser2 = {Email: "abc2", IsInvited: true, RemoteAccount: {Name: "xyz2"}};
		let invitedUser3 = {Email: "abc3", IsInvited: true, RemoteAccount: {Name: "xyz3"}};

		autotest(testdataInvited, `${__dirname}/testdata/DashboardUsers-InvitedUsers.json`,
			<DashboardUsers users={[invitedUser1, invitedUser2, invitedUser3]}
				currentUser={{UID: "1", Admin: false}}
				onboarding={{LinkGitHub: false}}
				allowStandaloneUsers={false}
				allowGitHubUsers={false} />
		);
	});
});

describe("DashboardUsers", () => {
	it("should render invited users", () => {
		let invitedUser1 = {Email: "abc1", IsInvited: false, RemoteAccount: {Name: "xyz1"}};
		let invitedUser2 = {Email: "abc2", IsInvited: false, RemoteAccount: {Name: "xyz2"}};
		let invitedUser3 = {Email: "abc3", IsInvited: false, RemoteAccount: {Name: "xyz3"}};

		autotest(testdataNotInvited, `${__dirname}/testdata/DashboardUsers-notInvited.json`,
			<DashboardUsers users={[invitedUser1, invitedUser2, invitedUser3]}
				currentUser={{UID: "1", Admin: false}}
				onboarding={{LinkGitHub: false}}
				allowStandaloneUsers={false}
				allowGitHubUsers={false} />
		);
	});
});