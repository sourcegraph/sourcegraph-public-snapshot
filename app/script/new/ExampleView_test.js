import autotest from "./util/autotest";

import React from "react";

import ExampleView from "./ExampleView";

import testdataInitial from "./testdata/ExampleView-initial.json";
import testdataAvailable from "./testdata/ExampleView-available.json";

describe("ExampleView", () => {
	it("should initially render empty and want example", () => {
		autotest(testdataInitial, `${__dirname}/testdata/ExampleView-initial.json`,
			<ExampleView
				defURL="/someURL"
				examples={{get(defURL, index) { return null; }, getCount(defURL) { return 10; }}}
				highlightedDef={null} />
		);
	});

	it("should display available example", () => {
		autotest(testdataAvailable, `${__dirname}/testdata/ExampleView-available.json`,
			<ExampleView
				defURL="/someURL"
				examples={{get(defURL, index) { return {Repo: "someRepo", File: "foo.go", StartLine: 3, EndLine: 7, SourceCode: {Lines: [{test: "aLine"}]}}; }, getCount(defURL) { return 10; }}}
				highlightedDef="/otherURL" />
		);
	});
});
