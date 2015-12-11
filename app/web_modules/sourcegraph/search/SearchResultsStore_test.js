import expect from "expect.js";

import Dispatcher from "../Dispatcher";
import SearchResultsStore from "./SearchResultsStore";
import * as SearchActions from "./SearchActions";

describe("SearchResultsStore", () => {
	it("should handle ResultsFetched", () => {
		Dispatcher.directDispatch(SearchResultsStore, new SearchActions.ResultsFetched("aRepo", "aRev", "aQuery", "aType", 1, "someResults"));
		expect(SearchResultsStore.results.get("aRepo", "aRev", "aQuery", "aType", 1)).to.be("someResults");
	});
});
