import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import {BuildStore} from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

let buildStore;
beforeEach(() => {
	buildStore = new BuildStore(Dispatcher);
});

describe("BuildStore", () => {
	it("should handle BuildFetched", () => {
		Dispatcher.directDispatch(buildStore, new BuildActions.BuildFetched("aRepo", 123, {ID: 123}));
		expect(buildStore.builds.get("aRepo", 123)).to.eql({ID: 123});
	});
});

describe("BuildStore", () => {
	it("should handle LogFetched", () => {
		Dispatcher.directDispatch(buildStore, new BuildActions.LogFetched("aRepo", 123, 456, 12, 34, "a"));
		expect(buildStore.logs.get("aRepo", 123, 456)).to.eql({maxID: 34, log: "a"});
	});

	it("should append cleanly to the log when handling a sequential LogFetched", () => {
		Dispatcher.directDispatch(buildStore, new BuildActions.LogFetched("bRepo", 123, 456, 0, 34, "a"));
		Dispatcher.directDispatch(buildStore, new BuildActions.LogFetched("bRepo", 123, 456, 34, 56, "b"));
		expect(buildStore.logs.get("bRepo", 123, 456)).to.eql({maxID: 56, log: "ab"});
	});
});

describe("BuildStore", () => {
	it("should handle TasksFetched", () => {
		Dispatcher.directDispatch(buildStore, new BuildActions.TasksFetched("aRepo", 123, [{ID: 456}]));
		expect(buildStore.tasks.get("aRepo", 123)).to.eql([{ID: 456}]);
	});
});
