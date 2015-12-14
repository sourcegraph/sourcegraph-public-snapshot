import autotest from "sourcegraph/util/autotest";

import React from "react";

import SearchFrameResults from "sourcegraph/search/SearchFrameResults";

import testdataNoResults from "./testdata/SearchFrameResults-noResults.json";
import testdataError from "./testdata/SearchFrameResults-error.json";
import testdataResultsGiven from "./testdata/SearchFrameResults-resultsGiven.json";


describe("SearchFrameResults", () => {
	it("should render an error when supplied", () => {
		autotest(testdataError, `${__dirname}/testdata/SearchFrameResults-error.json`,
			<SearchFrameResults resultData={{Error: "foo"}} repo="aRepo" rev="aRev" query="aQuery" page={1} label="aSearchFrame" />
		);
	});
});

describe("SearchFrameResults", () => {
	it("should render no results properly", () => {
		autotest(testdataNoResults, `${__dirname}/testdata/SearchFrameResults-noResults.json`,
			<SearchFrameResults resultData={{HTML: "<div>foo</div>", Total: 0}} repo="aRepo" rev="aRev" query="aQuery" page={1} label="aSearchFrame"/>
		);
	});
});

describe("SearchFrameResults", () => {
	it("should render results properly", () => {
		autotest(testdataResultsGiven, `${__dirname}/testdata/SearchFrameResults-resultsGiven.json`,
			<SearchFrameResults resultData={{HTML: "<div>foo</div>", Total: 9}} repo="aRepo" rev="aRev" query="aQuery" page={1} label="aSearchFrame"/>
		);
	});
});
