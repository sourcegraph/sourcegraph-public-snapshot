import autotest from "sourcegraph/util/autotest";
import React from "react";

import SearchBar from "sourcegraph/search/SearchBar";

import testdataRepo from "sourcegraph/search/testdata/SearchBar-repo.json";
import testdataSearchView from "sourcegraph/search/testdata/SearchBar-searchView.json";

describe("SearchBar", () => {
	it("should render inside a repo", () => {
		autotest(testdataRepo, `${__dirname}/testdata/SearchBar-repo.json`,
			<SearchBar location="http://localhost:3080/github.com/gorilla/mux@master" />
		);
	});

	it("should render an active search", () => {
		autotest(testdataSearchView, `${__dirname}/testdata/SearchBar-searchView.json`,
			<SearchBar location="http://localhost:3080/github.com/gorilla/mux@master/.search?q=test&type=tokens&page=1" />
		);
	});
});
