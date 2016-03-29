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
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("/someURL"));
			})).to.eql([new DefActions.DefFetched("/someURL", "someDefData")]);
		});

		it("with def not available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.api/repos/someURL");
				callback(null, {statusCode: 404}, null);
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantDef("/someURL"));
			})).to.eql([new DefActions.DefFetched("/someURL", {Error: true})]);
		});
	});

	it("should handle WantDefs", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.api/.defs?RepoRevs=myrepo@myrev&Nonlocal=true&Query=myquery");
			callback(null, {statusCode: 200}, {Defs: ["someDefData"]});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			DefBackend.__onDispatch(new DefActions.WantDefs("myrepo", "myrev", "myquery"));
		})).to.eql([new DefActions.DefsFetched("myrepo", "myrev", "myquery", {Defs: ["someDefData"]})]);
	});

	describe("should handle WantRefs", () => {
		it("for all files", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.ui/someURL/refs");
				callback(null, null, ["someRefData"]);
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("/someURL"));
			})).to.eql([new DefActions.RefsFetched("/someURL", null, ["someRefData"])]);
		});
		it("for a specific file", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/.ui/someURL/refs?Files=f");
				callback(null, {statusCode: 200}, ["someRefData"]);
			};
			expect(Dispatcher.Stores.catchDispatched(() => {
				DefBackend.__onDispatch(new DefActions.WantRefs("/someURL", "f"));
			})).to.eql([new DefActions.RefsFetched("/someURL", "f", ["someRefData"])]);
		});
	});
});
