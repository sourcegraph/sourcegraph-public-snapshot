import expect from "expect.js";

import Dispatcher from "../Dispatcher";
import SearchBackend from "./SearchBackend";
import * as SearchActions from "./SearchActions";

describe("SearchBackend", () => {
	it("should handle WantResults", () => {
		let search = {
			repo: "aRepo",
			rev: "aRev",
			type: "text",
			query: "aQuery",
			perPage: 15,
			page: 7,
		};
		let expectedURI = `/.ui/${search.repo}@${search.rev}/.search/${search.type}?q=${search.query}&PerPage=${search.perPage}&Page=${search.page}`;

		SearchBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, {Total: 42, Results: "someSearchResults"});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(SearchBackend, new SearchActions.WantResults(search.repo, search.rev, search.type, search.page, search.perPage, search.query));
		})).to.eql([new SearchActions.ResultsFetched(search.repo, search.rev, search.query, search.type, search.page, {Total: 42, Results: "someSearchResults"})]);
	});
});
