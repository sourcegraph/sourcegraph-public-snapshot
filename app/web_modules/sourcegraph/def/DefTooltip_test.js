import autotest from "sourcegraph/util/autotest";

import React from "react";

import DefTooltip from "sourcegraph/def/DefTooltip";

import testdataData from "sourcegraph/def/testdata/DefTooltip-data.json";

describe("DefTooltip", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefTooltip-data.json`,
			<DefTooltip def={{URL: "someURL", QualifiedName: {__html: "someName"}, Data: {DocHTML: "someDoc", Repo: "someRepo"}}} />
		);
	});
});
