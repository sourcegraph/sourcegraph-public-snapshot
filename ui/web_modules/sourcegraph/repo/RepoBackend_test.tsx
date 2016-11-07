import expect from "expect.js";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoBackend} from "sourcegraph/repo/RepoBackend";
import {immediateSyncPromise} from "sourcegraph/util/testutil/immediateSyncPromise";

describe("RepoBackend", () => {
	it("should handle WantCreateRepo for mirror repo", () => {
		RepoBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/repos");
			expect(init.method).to.be("POST");
			expect(JSON.parse(init.body as string)).to.eql({Op: {Origin: {ID: "123", Service: 0}}});
			return immediateSyncPromise({status: 200, json: () => ({ID: 1})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			RepoBackend.__onDispatch(new RepoActions.WantCreateRepo("r", {GitHubID: 123}));
		})).to.eql([new RepoActions.RepoCreated("r", {ID: 1})]);
	});
	it("should handle WantCreateRepo for non-mirror repo", () => {
		RepoBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/repos");
			expect(init.method).to.be("POST");
			expect(JSON.parse(init.body as string)).to.eql({
				Op: {New: {
					URI: "a/b",
					CloneURL: "https://a/b",
					DefaultBranch: "master",
					Mirror: true,
				}},
			});
			return immediateSyncPromise({status: 200, json: () => ({ID: 1})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			RepoBackend.__onDispatch(new RepoActions.WantCreateRepo("a/b", {HTTPCloneURL: "https://a/b"}));
		})).to.eql([new RepoActions.RepoCreated("a/b", {ID: 1})]);
	});
});
