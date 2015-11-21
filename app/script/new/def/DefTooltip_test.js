import autotest from "../util/autotest";

import React from "react";

import DefTooltip from "./DefTooltip";

import testdataData from "./testdata/DefTooltip-data.json";

describe("DefTooltip", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefTooltip-data.json`,
			<DefTooltip def={{URL: "someURL", QualifiedName: {__html: "someName"}, Data: {DocHTML: "someDoc", Repo: "someRepo"}}} />
		);
	});
});
