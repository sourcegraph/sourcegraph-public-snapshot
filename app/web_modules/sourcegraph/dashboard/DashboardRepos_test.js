import autotest from "sourcegraph/util/autotest";

import React from "react";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import {withUserContext} from "sourcegraph/app/user";
import testdataData from "sourcegraph/dashboard/testdata/DashboardRepos-data.json";

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
			React.createElement(withUserContext(<DashboardRepos
				repos={repos}
				exampleRepos={repos}
				linkGitHubURL={""} />)),
			{signedIn: false},
		);
	});
});
