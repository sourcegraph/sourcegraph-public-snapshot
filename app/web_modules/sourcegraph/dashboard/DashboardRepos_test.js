import autotest from "sourcegraph/util/autotest";

import React from "react";
import moment from "moment";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import testdataData from "sourcegraph/dashboard/testdata/DashboardRepos-data.json";
import testdataNotSupported from "sourcegraph/dashboard/testdata/DashboardRepos-notSupported.json";
import testdataOnWaitlist from "sourcegraph/dashboard/testdata/DashboardRepos-onWaitlist.json";

describe("DashboardRepos", () => {
	it("should render repos", () => {
		autotest(testdataData, `${__dirname}/testdata/DashboardRepos-data.json`,
			<DashboardRepos
				repos={[{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment(), Language: "Go"}]}
				remoteRepos={[{Private: false, Owner: "o", Name: "r"}]}
				onWaitlist={false}
				allowGitHubMirrors={true}
				linkGitHub={false} />
		);
	});
});

describe("DashboardRepos", () => {
	it("should render repos", () => {
		let repos = [
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment(), Language: "C++"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(1, "seconds"), Language: "C"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(2, "seconds"), Language: "Python"}];
		autotest(testdataNotSupported, `${__dirname}/testdata/DashboardRepos-notSupported.json`,
			<DashboardRepos
				repos={repos}
				remoteRepos={[{Private: false, Owner: "o", Name: "r", Language: "C++"}]}
				onWaitlist={false}
				allowGitHubMirrors={true}
				linkGitHub={false} />
		);
	});
});

describe("DashboardRepos", () => {
	it("should render repos on waitlist", () => {
		let repos = [
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment(), Language: "C++"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(1, "seconds"), Language: "C"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(2, "seconds"), Language: "Python"}];
		autotest(testdataOnWaitlist, `${__dirname}/testdata/DashboardRepos-onWaitlist.json`,
			<DashboardRepos
				repos={repos}
				remoteRepos={[{Private: false, Owner: "o", Name: "r", Language: "C++"}]}
				onWaitlist={true}
				allowGitHubMirrors={true}
				linkGitHub={false} />
		);
	});
});
