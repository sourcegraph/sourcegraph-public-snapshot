import autotest from "sourcegraph/util/autotest";

import React from "react";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import testdataData from "sourcegraph/dashboard/testdata/DashboardRepos-data.json";
import testdataUnsupported from "sourcegraph/dashboard/testdata/DashboardRepos-unsupported.json";

describe("DashboardRepos", () => {
	it("should render repos", () => {
		let repos=[{
			Private: false,
			URI: "someURL",
			Description: "someDescription",
			UpdatedAt: "2016-02-24T10:18:55-08:00",
			Language: "Go",
		}];
		autotest(testdataData, `${__dirname}/testdata/DashboardRepos-data.json`,
			<DashboardRepos
				repos={repos}
				exampleRepos={repos}
				hasLinkedGitHub={true}
				linkGitHubURL={""} />
		);
	});
});

describe("DashboardRepos", () => {
	it("should render unsupported repos", () => {
		let repos = [
			{GitHubID: 1, Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: new Date(4).toISOString(), Language: "C++"},
			{GitHubID: 2, Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: new Date(3).toISOString(), Language: "C"},
			{GitHubID: 3, Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: new Date(2).toISOString(), Language: "Python"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: new Date(1).toISOString(), Language: "Scala"},
		];
		autotest(testdataUnsupported, `${__dirname}/testdata/DashboardRepos-unsupported.json`,
			<DashboardRepos
				repos={repos}
				exampleRepos={repos}
				hasLinkedGitHub={true}
				linkGitHubURL={""} />
		);
	});
});
