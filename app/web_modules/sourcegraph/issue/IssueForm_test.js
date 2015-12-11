import autotest from "sourcegraph/util/autotest";

import React from "react";

import IssueForm from "sourcegraph/issue/IssueForm";

import testdataForm from "sourcegraph/issue/testdata/IssueForm-form.json";

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
