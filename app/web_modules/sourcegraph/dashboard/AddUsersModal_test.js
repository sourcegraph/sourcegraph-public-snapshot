import autotest from "sourcegraph/util/autotest";

import React from "react";
import moment from "moment";
import AddUsersModal from "sourcegraph/dashboard/AddUsersModal";

import testdataData from "sourcegraph/dashboard/testdata/AddUsersModal-data.json";

describe("AddUsersModal", () => {
	it("should render Add Users Modal", () => {
		autotest(testdataData, `${__dirname}/testdata/AddUsersModal-data.json`,
			<AddUsersModal allowStandaloneUsers={false} allowGitHubUsers={true} />
		);
	});
});
