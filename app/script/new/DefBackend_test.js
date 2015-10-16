import expect from "expect.js";

import Dispatcher from "./Dispatcher";
import DefBackend from "./DefBackend";
import * as DefActions from "./DefActions";

describe("DefBackend", () => {
	it("should handle WantDef", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/someURL");
			expect(options.headers).to.eql({"X-Definition-Data-Only": "yes"});
			callback(null, null, {Found: true, Data: "someDefData"});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/someURL"));
		})).to.eql([new DefActions.DefFetched("/someURL", {Found: true, Data: "someDefData"})]);
	});

	it("should not forward unavailable definition", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/otherURL");
			expect(options.headers).to.eql({"X-Definition-Data-Only": "yes"});
			callback(null, null, {Found: false});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/otherURL"));
		})).to.eql([]);
	});
});
