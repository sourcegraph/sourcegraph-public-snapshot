import autotest from "../util/autotest";

import React from "react";

import TextSearchResults from "./TextSearchResults";

import testdataNoResults from "./testdata/TextSearchResults-noResults.json";
import testdataResults from "./testdata/TextSearchResults-results.json";

describe("TextSearchResults", () => {
	it("should render no results", () => {
		autotest(testdataNoResults, `${__dirname}/testdata/TextSearchResults-noResults.json`,
			<TextSearchResults resultData={{Results: [], Total: 0}} repo="aRepo" rev="aRev" query="aQuery" page={1} />
		);
	});

	it("should render results", () => {
		let exampleResult = {
			File: "app.go",
			StartLine: 1,
			EndLine: 10,
			Lines: [],
		};
		autotest(testdataResults, `${__dirname}/testdata/TextSearchResults-results.json`,
			<TextSearchResults resultData={{Results: [exampleResult], Total: 1}} repo="aRepo" rev="aRev" query="aQuery" page={1} />
		);
	});
});
