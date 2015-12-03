import autotest from "../util/autotest";

import React from "react";

import SearchResultsContainer from "./SearchResultsContainer";
import SearchResultsStore from "./SearchResultsStore";
import * as SearchActions from "./SearchActions";
import Dispatcher from "../Dispatcher";

import testdataUnfetched from "./testdata/SearchResultsContainer-unfetched.json";
import testdataFetched from "./testdata/SearchResultsContainer-fetched.json";

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
