import * as React from "react";
import {SignedInDashboard} from "sourcegraph/dashboard/SignedInDashboard";
import testdataData from "sourcegraph/dashboard/testdata/SignedInDashboard-data.json";
import {autotest} from "sourcegraph/util/testutil/autotest";

describe("SignedInDashboard", () => {
	it("should render a dashboard", () => {
		autotest(testdataData, "sourcegraph/dashboard/testdata/SignedInDashboard-data.json",
			<SignedInDashboard location={{}} />,
			{
				siteConfig: {},
				router: {push: () => { /* ignore */ }},
			},
		);
	});
});
