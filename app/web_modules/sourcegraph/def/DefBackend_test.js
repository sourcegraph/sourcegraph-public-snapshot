import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefBackend from "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefBackend", () => {
	describe("should handle WantDef", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.ui/someURL");
			expect(options.headers).to.eql({"X-Definition-Data-Only": "yes"});
			callback(null, null, {Found: true, Data: "someDefData"});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/someURL"));
		})).to.eql([new DefActions.DefFetched("/someURL", {Found: true, Data: "someDefData"})]);
	});


	describe("should handle WantExample", () => {
		it("with result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.ui/someURL/.examples?TokenizedSource=true&PerPage=1&Page=43");
				callback(null, null, [{test: "exampleData"}]);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.ExampleFetched("/someURL", 42, {test: "exampleData"})]);
		});

		it("with no result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.ui/someURL/.examples?TokenizedSource=true&PerPage=1&Page=43");
				callback(null, null, null);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.NoExampleAvailable("/someURL", 42)]);
		});
	});
});
