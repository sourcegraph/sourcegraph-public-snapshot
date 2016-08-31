import * as React from "react";
import {withUserContext} from "sourcegraph/app/user";
import {DashboardContainer} from "sourcegraph/dashboard/DashboardContainer";
import testdataData from "sourcegraph/dashboard/testdata/DashboardContainer-data.json";
import {autotest} from "sourcegraph/util/testutil/autotest";

describe("DashboardContainer", () => {
	it("should render a dashboard", () => {
		autotest(testdataData, "sourcegraph/dashboard/testdata/DashboardContainer-data.json",
			React.createElement(withUserContext(<DashboardContainer />)),
			{
				siteConfig: {},
				signedIn: false,
				user: null,
				githubToken: null,
				eventLogger: {logEvent: () => null},
			},
		);
	});
});
