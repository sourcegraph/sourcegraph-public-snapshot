import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefBackend from "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
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
			})).to.eql([new DefActions.DefFetched("r", "v", "d", "someDefData")]);
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
			expect(url).to.be("/.api/defs?RepoRevs=myrepo@myrev&Nonlocal=true&Query=myquery");
			return immediateSyncPromise({status: 200, json: () => ({Defs: ["someDefData"]})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DefBackend.__onDispatch(new DefActions.WantDefs("myrepo", "myrev", "myquery"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "myrev", "myquery", {Defs: ["someDefData"]})]);
	});

	describe("should handle WantRefs", () => {
		it("for all files", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", null, "someRefData")]);
		});
		it("for a specific file", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Files=f");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "f"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "f", "someRefData")]);
		});
		it("with no result available", () => {
			DefBackend.fetch = function(url, options) {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs");
				return immediateSyncPromise({status: 200, json: () => null});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", null, null)]);
		});
	});
});
