import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefBackend from "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";

describe("DefBackend", () => {
	describe("should handle WantDef", () => {
		it("with def available", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d");
				return immediateSyncPromise({status: 200, json: () => "someDefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("r", "v", "d"));
			})).to.eql([
				new RepoActions.RepoCloning("r", false),
				new DefActions.DefFetched("r", "v", "d", "someDefData"),
			]);
		});

		it("with def not available", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d");
				return immediateSyncPromise({status: 404, json: () => null});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("r", "v", "d"));
			})).to.eql([new DefActions.DefFetched("r", "v", "d", {Error: true})]);
		});
	});

	it("should handle WantDefs", () => {
		DefBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/defs?RepoRevs=myrepo@mycommitID&Nonlocal=true&Query=myquery&FilePathPrefix=foo%2F");
			return immediateSyncPromise({status: 200, json: () => ({Defs: ["someDefData"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DefBackend.__onDispatch(new DefActions.WantDefs("myrepo", "mycommitID", "myquery", "foo/"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "mycommitID", "myquery", "foo/", {Defs: ["someDefData"]})]);
	});

	it("should handle WantRefLocations", () => {
		DefBackend.fetch = function(url, options) {
			expect(url).to.contain("/.api/repos/r@v/-/def/d/-/ref-locations");
			return immediateSyncPromise({status: 200, json: () => "someRefData"});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DefBackend.__onDispatch(new DefActions.WantRefLocations({
				repo: "r", commitID: "v", def: "d", repos: [],
			}));
		})).to.eql([new DefActions.RefLocationsFetched(
			new DefActions.WantRefLocations({
				repo: "r", commitID: "v", def: "d", repos: [],
			}), "someRefData")]);
	});

	describe("should handle WantRefs", () => {
		it("for all files", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", null, "someRefData")]);
		});
		it("for a specific file", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr&Files=rf");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr", "rf"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", "rf", "someRefData")]);
		});
		it("with no result available", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr");
				return immediateSyncPromise({status: 200, json: () => null});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", null, null)]);
		});
	});
});
