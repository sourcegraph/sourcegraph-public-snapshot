import autotest from "sourcegraph/util/autotest";

import React from "react";
import moment from "moment";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import testdataData from "sourcegraph/dashboard/testdata/DashboardRepos-data.json";
import testdataUnsupported from "sourcegraph/dashboard/testdata/DashboardRepos-unsupported.json";

describe("DashboardRepos", () => {
	it("should render repos", () => {
		autotest(testdataData, `${__dirname}/testdata/DashboardRepos-data.json`,
			<DashboardRepos
				repos={[{
					Private: false,
					URI: "someURL",
					Description: "someDescription",
					UpdatedAt: moment(),
					Language: "Go",
				}]}
				linkGitHub={false} />
		);
	});
});

describe("DashboardRepos", () => {
	it("should render unsupported repos", () => {
		let repos = [
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment(), Language: "C++"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(1, "seconds"), Language: "C"},
			{Private: false, URI: "someURL", Description: "someDescription", UpdatedAt: moment().subtract(2, "seconds"), Language: "Python"},
		];
		autotest(testdataUnsupported, `${__dirname}/testdata/DashboardRepos-unsupported.json`,
			<DashboardRepos
				repos={repos}
				linkGitHub={false} />
		);
	});
});
