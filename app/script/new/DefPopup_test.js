import autotest from "./util/autotest";

import React from "react";

import DefPopup from "./DefPopup";

import testdataData from "./testdata/DefPopup-data.json";
import testdataNotAvailable from "./testdata/DefPopup-notAvailable.json";

describe("DefPopup", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefPopup-data.json`,
			<DefPopup
				def={{Found: true, URL: "someURL", QualifiedName: {__html: "someName"}, Data: {DocHTML: "someDoc"}}}
				examples={{test: "examples", generation: 42}}
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

	it("should render 'not available'", () => {
		autotest(testdataNotAvailable, `${__dirname}/testdata/DefPopup-notAvailable.json`,
			<DefPopup
				def={{Found: false, URL: "someURL", QualifiedName: {__html: "someName"}}}
				examples={{test: "examples", generation: 42}}
				discussions={[]}
				highlightedDef={null} />
		);
	});
});
