import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import DefBackend from "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";

describe("DefBackend", () => {
	describe("should handle WantDef", () => {
		it("with def available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL");
				callback(null, {statusCode: 200}, "someDefData");
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/someURL"));
			})).to.eql([new DefActions.DefFetched("/someURL", "someDefData")]);
		});

		it("with def not available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL");
				callback(null, {statusCode: 404}, null);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantDef("/someURL"));
			})).to.eql([new DefActions.DefFetched("/someURL", {Error: true})]);
		});
	});

	it("should handle WantDefs", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.api/.defs?RepoRevs=myrepo@myrev&Nonlocal=true&Query=myquery");
			callback(null, {statusCode: 200}, {Defs: ["someDefData"]});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDefs("myrepo", "myrev", "myquery"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "myrev", "myquery", {Defs: ["someDefData"]})]);
	});

	describe("should handle WantExamples", () => {
		it("with result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL/.examples?PerPage=10&Page=42");
				callback(null, {statusCode: 200}, {Examples: [{test: "exampleData"}]});
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExamples("/someURL", 42));
			})).to.eql([new DefActions.ExamplesFetched("/someURL", 42, {Examples: [{test: "exampleData"}]})]);
		});

		it("with no result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL/.examples?PerPage=10&Page=42");
				callback(null, {statusCode: 200}, null);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExamples("/someURL", 42));
			})).to.eql([new DefActions.NoExamplesAvailable("/someURL", 42)]);
		});
	});
});
