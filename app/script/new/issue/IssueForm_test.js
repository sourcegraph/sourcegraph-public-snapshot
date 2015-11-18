import autotest from "../util/autotest";

import React from "react";

import IssueForm from "./IssueForm";

import testdataForm from "./testdata/IssueForm-form.json";

describe("IssueForm", () => {
	it("should render form", () => {
		autotest(testdataForm, `${__dirname}/testdata/IssueForm-form.json`,
			<IssueForm
				repo="aRepo"
				path="a/path"
				commitID={"a".repeat(40)}
				startLine={1}
				endLine={42}
				onSubmit={function() {}} />
		);
	});
});
