import expect from "expect.js";
import * as DefActions from "sourcegraph/def/DefActions";
import {DefBackend} from "sourcegraph/def/DefBackend";
import {Def, Ref} from "sourcegraph/def/index";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {immediateSyncPromise} from "sourcegraph/util/immediateSyncPromise";

describe("DefBackend", () => {
	describe("should handle WantDef", () => {
		it("with def available", () => {
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d");
				return immediateSyncPromise({status: 200, json: () => ({Path: "somePath"})});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("r", "v", "d"));
			})).to.eql([
				new RepoActions.RepoCloning("r", false),
				new DefActions.DefFetched("r", "v", "d", {Path: "somePath"} as Def),
			]);
		});

		it("with def not available", () => {
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d");
				return immediateSyncPromise({status: 404, json: () => null});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("r", "v", "d"));
			})).to.eql([new DefActions.DefFetched("r", "v", "d", {Error: true} as Def)]);
		});
	});

	it("should handle WantDefs", () => {
		DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/defs?RepoRevs=myrepo@mycommitID&Nonlocal=true&Query=myquery&FilePathPrefix=foo%2F");
			return immediateSyncPromise({status: 200, json: () => ([{Path: "somePath"}])});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DefBackend.__onDispatch(new DefActions.WantDefs("myrepo", "mycommitID", "myquery", "foo/"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "mycommitID", "myquery", "foo/", [{Path: "somePath"} as Def])]);
	});

	it("should handle WantRefLocations", () => {
		DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
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
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", null, "someRefData" as any as Ref[])]);
		});
		it("for a specific file", () => {
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr&Files=rf");
				return immediateSyncPromise({status: 200, json: () => "someRefData"});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr", "rf"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", "rf", "someRefData" as any as Ref[])]);
		});
		it("with no result available", () => {
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr");
				return immediateSyncPromise({status: 200, json: () => null});
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", null, null)]);
		});
		it("for a 404 error", () => {
			DefBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
				expect(url).to.be("/.api/repos/r@v/-/def/d/-/refs?Repo=rr");
				return immediateSyncPromise({response: {status: 404}, text: ""}, true);
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("r", "v", "d", "rr"));
			})).to.eql([new DefActions.RefsFetched("r", "v", "d", "rr", null, {} as any)]);
		});
	});
});
