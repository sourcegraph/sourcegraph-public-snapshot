import autotest from "sourcegraph/util/autotest";

import React from "react";

import SearchResultsContainer from "sourcegraph/search/SearchResultsContainer";
import SearchResultsStore from "sourcegraph/search/SearchResultsStore";
import * as SearchActions from "sourcegraph/search/SearchActions";
import Dispatcher from "sourcegraph/Dispatcher";

import testdataUnfetched from "sourcegraph/search/testdata/SearchResultsContainer-unfetched.json";
import testdataFetched from "sourcegraph/search/testdata/SearchResultsContainer-fetched.json";

describe("SearchResultsContainer", () => {
	it("should handle unfetched results", () => {
		autotest(testdataUnfetched, `${__dirname}/testdata/SearchResultsContainer-unfetched.json`,
			<SearchResultsContainer repo="aRepo" rev="aRev" query="foo" type="tokens" page={1} />
		);
	});

	it("should display fetched results", () => {
		let results = {
			Total: 42,
		};
		Dispatcher.directDispatch(SearchResultsStore, new SearchActions.ResultsFetched(
			"aRepo", "aRev", "foo", "text", 1, results
		));
		autotest(testdataFetched, `${__dirname}/testdata/SearchResultsContainer-fetched.json`,
			<SearchResultsContainer repo="aRepo" rev="aRev" query="foo" type="text" page={1}/>
		);
	});
});
