// tslint:disable

import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import RepoBackend from "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";

describe("RepoBackend", () => {
	it("should handle WantCreateRepo for mirror repo", () => {
		RepoBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos");
			expect(options.method).to.be("POST");
			expect(JSON.parse(options.body)).to.eql({Op: {Origin: {ID: "123", Service: 0}}});
			return immediateSyncPromise({status: 200, json: () => ({ID: 1})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			RepoBackend.__onDispatch(new RepoActions.WantCreateRepo("r", {GitHubID: 123}));
		})).to.eql([new RepoActions.RepoCreated("r", {ID: 1})]);
	});
	it("should handle WantCreateRepo for non-mirror repo", () => {
		RepoBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos");
			expect(options.method).to.be("POST");
			expect(JSON.parse(options.body)).to.eql({
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
	it("should handle WantCommit", () => {
		RepoBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos/r@v/-/commit");
			return immediateSyncPromise({status: 200, json: () => ({ID: "c"})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			RepoBackend.__onDispatch(new RepoActions.WantCommit("r", "v"));
		})).to.eql([new RepoActions.FetchedCommit("r", "v", {ID: "c"})]);
	});
});
