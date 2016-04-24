import autotest from "sourcegraph/util/autotest";

import React from "react";
import DashboardContainer from "sourcegraph/dashboard/DashboardContainer";

import testdataData from "sourcegraph/dashboard/testdata/DashboardContainer-data.json";

describe("DashboardContainer", () => {
	it("should render a dashboard", () => {
		autotest(testdataData, `${__dirname}/testdata/DashboardContainer-data.json`,
			<DashboardContainer />,
			{siteConfig: {}},
		);
	});
});
