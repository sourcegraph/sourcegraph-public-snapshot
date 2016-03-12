import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefBackend from "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefBackend", () => {
	it("should handle WantDef", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.ui/someURL");
			expect(options.headers).to.eql({"X-Definition-Data-Only": "yes"});
			callback(null, null, {Data: "someDefData"});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/someURL"));
		})).to.eql([new DefActions.DefFetched("/someURL", {Data: "someDefData"})]);
	});

	it("should handle WantDefs", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.api/.defs?RepoRevs=myrepo@myrev&Nonlocal=true&Query=myquery");
			callback(null, null, {Defs: ["someDefData"]});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDefs("myrepo", "myrev", "myquery"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "myrev", "myquery", {Defs: ["someDefData"]})]);
	});

	describe("should handle WantExample", () => {
		it("with result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL/.examples?PerPage=1&Page=43");
				callback(null, {statusCode: 200}, {Examples: [{test: "exampleData"}]});
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.ExampleFetched("/someURL", 42, {test: "exampleData"})]);
		});

		it("with no result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL/.examples?PerPage=1&Page=43");
				callback(null, {statusCode: 200}, null);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.NoExampleAvailable("/someURL", 42)]);
		});
	});
});
