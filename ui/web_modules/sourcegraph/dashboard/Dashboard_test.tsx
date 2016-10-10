import * as React from "react";
import {Dashboard} from "sourcegraph/dashboard/Dashboard";
import testdataData from "sourcegraph/dashboard/testdata/Dashboard-data.json";
import {autotest} from "sourcegraph/util/testutil/autotest";

describe("Dashboard", () => {
	it("should render a dashboard", () => {
		autotest(testdataData, "sourcegraph/dashboard/testdata/Dashboard-data.json",
			<Dashboard location={{
				pathname: "foo",
				action: "bar",
				search: "foo",
				hash: "",
				query:{},
				state:{},
				key: "foo",
			}} />
		);
	});
});
