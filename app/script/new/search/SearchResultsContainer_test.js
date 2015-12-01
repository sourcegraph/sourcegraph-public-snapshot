import autotest from "../util/autotest";

import React from "react";

import SearchResultsContainer from "./SearchResultsContainer";
import SearchResultsStore from "./SearchResultsStore";
import * as SearchActions from "./SearchActions";
import Dispatcher from "../Dispatcher";

import testdataUnfetched from "./testdata/SearchResultsContainer-unfetched.json";
import testdataFetched from "./testdata/SearchResultsContainer-fetched.json";

describe("SearchResultsContainer", () => {
	let exampleResults = {
		Total: 42,
	};

	let exampleRepo = {
		repo: "aRepo",
		rev: "aRev",
	};

	it("should handle unfetched results", () => {
		autotest(testdataUnfetched, `${__dirname}/testdata/SearchResultsContainer-unfetched.json`,
			<SearchResultsContainer type="tokens" page={1} query="foo" {...exampleRepo} />
		);
	});

	it("should display fetched results", () => {
		let query = "foo";
		let type = "text";
		let page = 1;
		Dispatcher.directDispatch(SearchResultsStore, new SearchActions.ResultsFetched(
			exampleRepo.repo, exampleRepo.rev, query, type, page, exampleResults
		));
		autotest(testdataFetched, `${__dirname}/testdata/SearchResultsContainer-fetched.json`,
			<SearchResultsContainer query={query} type={type} page={page} {...exampleRepo} />
		);
	});
});
