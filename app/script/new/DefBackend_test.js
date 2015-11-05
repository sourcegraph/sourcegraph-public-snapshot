import expect from "expect.js";

import Dispatcher from "./Dispatcher";
import DefBackend from "./DefBackend";
import DefStore from "./DefStore";
import * as DefActions from "./DefActions";

describe("DefBackend", () => {
	describe("should handle WantDef", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/someURL");
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
				expect(options.uri).to.be("/ui/someURL/.examples?TokenizedSource=true&PerPage=1&Page=43");
				callback(null, null, [{test: "exampleData"}]);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.ExampleFetched("/someURL", 42, {test: "exampleData"})]);
		});

		it("with no result available", () => {
			DefBackend.xhr = function(options, callback) {
				expect(options.uri).to.be("/ui/someURL/.examples?TokenizedSource=true&PerPage=1&Page=43");
				callback(null, null, null);
			};
			expect(Dispatcher.catchDispatched(() => {
				Dispatcher.directDispatch(DefBackend, new DefActions.WantExample("/someURL", 42));
			})).to.eql([new DefActions.NoExampleAvailable("/someURL", 42)]);
		});
	});

	describe("should handle WantDiscussions", () => {
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/someURL/.discussions?order=Date");
			callback(null, null, {Discussions: [{ID: 42, Comments: []}]});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(DefBackend, new DefActions.WantDiscussions("/someURL", 42));
		})).to.eql([new DefActions.DiscussionsFetched("/someURL", [{ID: 42, Comments: []}])]);
	});

	describe("should handle CreateDiscussion", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DiscussionsFetched("/someURL", [{ID: 42, Comments: []}]));
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/someURL/.discussions/create");
			expect(options.method).to.be("POST");
			expect(options.json).to.eql({Title: "someTitle", Description: "someDescription"});
			callback(null, null, {ID: 43, Comments: []});
		};
		expect(Dispatcher.catchDispatched(() => {
			let callbackDiscussion;
			Dispatcher.directDispatch(DefBackend, new DefActions.CreateDiscussion("/someURL", "someTitle", "someDescription", function(d) { callbackDiscussion = d; }));
			expect(callbackDiscussion).to.eql({ID: 43, Comments: []});
		})).to.eql([new DefActions.DiscussionsFetched("/someURL", [{ID: 43, Comments: []}, {ID: 42, Comments: []}])]);
	});

	describe("should handle CreateDiscussionComment", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DiscussionsFetched("/someURL", [{ID: 42, Comments: [{ID: 0}]}]));
		DefBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/someURL/.discussions/42/.comment");
			expect(options.method).to.be("POST");
			expect(options.json).to.eql({Body: "someBody"});
			callback(null, null, {ID: 1});
		};
		expect(Dispatcher.catchDispatched(() => {
			let callbackCalled;
			Dispatcher.directDispatch(DefBackend, new DefActions.CreateDiscussionComment("/someURL", 42, "someBody", function() { callbackCalled = true; }));
			expect(callbackCalled).to.be(true);
		})).to.eql([new DefActions.DiscussionsFetched("/someURL", [{ID: 42, Comments: [{ID: 0}, {ID: 1}]}])]);
	});
});
