import autotest from "./util/autotest";

import React from "react";

import DefPopup from "./DefPopup";

import testdataData from "./testdata/DefPopup-data.json";
import testdataNoDiscussions from "./testdata/DefPopup-noDiscussions.json";

describe("DefPopup", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefPopup-data.json`,
			<DefPopup
				def={{URL: "someURL", QualifiedName: {__html: "someName"}, Data: {DocHTML: "someDoc"}}}
				examples={{test: "examples"}}
				discussions={[
					{ID: 0, Comments: []},
					{ID: 1, Comments: []},
					{ID: 2, Comments: [{ID: 0}]},
					{ID: 3, Comments: []},
					{ID: 4, Comments: []},
				]}
				highlightedDef="otherURL" />
		);
	});

	it("should render no discussions", () => {
		autotest(testdataNoDiscussions, `${__dirname}/testdata/DefPopup-noDiscussions.json`,
			<DefPopup
				def={{URL: "someURL", QualifiedName: {__html: "someName"}, Data: {DocHTML: "someDoc"}}}
				examples={{test: "examples"}}
				discussions={[]} />
		);
	});
});
