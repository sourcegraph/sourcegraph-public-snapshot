import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import SearchResultsStore from "sourcegraph/search/SearchResultsStore";
import * as SearchActions from "sourcegraph/search/SearchActions";

describe("SearchResultsStore", () => {
	it("should handle ResultsFetched", () => {
		Dispatcher.directDispatch(SearchResultsStore, new SearchActions.ResultsFetched("aRepo", "aRev", "aQuery", "aType", 1, "someResults"));
		expect(SearchResultsStore.results.get("aRepo", "aRev", "aQuery", "aType", 1)).to.be("someResults");
	});
});
