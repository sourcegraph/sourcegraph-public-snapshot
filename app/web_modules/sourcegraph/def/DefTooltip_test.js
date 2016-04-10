import autotest from "sourcegraph/util/autotest";

import React from "react";

import DefTooltip from "sourcegraph/def/DefTooltip";

import testdataData from "sourcegraph/def/testdata/DefTooltip-data.json";

const fmtStrings = {DefKeyword: "a", NameAndTypeSeparator: "s", Name: {ScopeQualified: "n"}, Type: {ScopeQualified: "t"}};

describe("DefTooltip", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefTooltip-data.json`,
			<DefTooltip def={{URL: "someURL", FmtStrings: fmtStrings, DocHTML: "someDoc", Repo: "someRepo"}} />
		);
	});
});
