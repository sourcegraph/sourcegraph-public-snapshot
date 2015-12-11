import autotest from "sourcegraph/util/autotest";
import expect from "expect.js";

import React from "react";
import TestUtils from "react-addons-test-utils";

import Dispatcher from "sourcegraph/Dispatcher";
import SearchResultsRouter from "sourcegraph/search/SearchResultsRouter";
import * as SearchActions from "sourcegraph/search/SearchActions";

import testdataSearch from "sourcegraph/search/testdata/SearchResultsRouter-search.json";

describe("SearchResultsRouter", () => {
	it("should handle search URLs", () => {
		autotest(testdataSearch, `${__dirname}/testdata/SearchResultsRouter-search.json`,
			<SearchResultsRouter location="http://localhost:3080/github.com/gorilla/mux@master/.search?q=foo&type=tokens&page=1" />
		);
	});

	it("should handle SearchActions.SelectResultType", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.search?q=foo&type=tokens&page=1",
			new SearchActions.SelectResultType("text"),
			"http://localhost:3080/github.com/gorilla/mux@master/.search?q=foo&type=text&page=1"
		);
	});

	it("should handle SearchActions.SelectPage", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.search?q=foo&type=text&page=1",
			new SearchActions.SelectPage(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.search?q=foo&type=text&page=42"
		);
	});
});

function testAction(uri, action, expectedURI) {
	let renderer = TestUtils.createRenderer();
	renderer.render(<SearchResultsRouter location={uri} navigate={(newURI) => { uri = newURI; }} />);
	Dispatcher.directDispatch(renderer._instance._instance, action);
	expect(uri).to.be(expectedURI);
}
