// tslint:disable

import autotest from "sourcegraph/util/autotest";

import * as React from "react";

import DefPopup from "sourcegraph/def/DefPopup";

import testdataData from "sourcegraph/def/testdata/DefPopup-data.json";
import testdataNotAvailable from "sourcegraph/def/testdata/DefPopup-notAvailable.json";

const fmtStrings = {DefKeyword: "a", NameAndTypeSeparator: "s", Name: {ScopeQualified: "n"}, Type: {ScopeQualified: "t"}};

describe("DefPopup", () => {
	it("should render definition data", () => {
		autotest(testdataData, "sourcegraph/def/testdata/DefPopup-data.json",
			<DefPopup
				def={{Repo: "r", CommitID: "c", FmtStrings: fmtStrings, DocHTML: "someDoc"}}
				examples={{test: "examples"}}
				highlightedDef="otherURL" />,
				{features: {}},
		);
	});

	it("should render 'not available'", () => {
		autotest(testdataNotAvailable, "sourcegraph/def/testdata/DefPopup-notAvailable.json",
			<DefPopup
				def={{Repo: "r", CommitID: "c", FmtStrings: fmtStrings}}
				examples={{test: "examples"}}
				highlightedDef={null} />,
				{features: {}},
		);
	});
});
