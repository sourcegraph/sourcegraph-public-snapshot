import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import CodeBackend from "sourcegraph/code/CodeBackend";
import {prepareAnnotations} from "sourcegraph/code/CodeBackend";
import * as CodeActions from "sourcegraph/code/CodeActions";

describe("CodeBackend", () => {
	it("should handle WantFile", () => {
		CodeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.ui/aRepo@aRev/.tree/aTree");
			callback(null, null, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(CodeBackend, new CodeActions.WantFile("aRepo", "aRev", "aTree"));
		})).to.eql([new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someFile")]);
	});
});

describe("prepareAnnotations", () => {
	it("should set WantInner on syntax highlighting annotations", () => {
		expect(
			prepareAnnotations([
				{StartByte: 1, Class: "x"},
				{StartByte: 2, URL: "y"},
				{StartByte: 3, Class: "z"},
			])
		).to.eql(
			[
				{StartByte: 1, Class: "x", WantInner: 1},
				{StartByte: 2, URL: "y"},
				{StartByte: 3, Class: "z", WantInner: 1},
			]
		);
	});
	it("should combine coincident ref annotations (for multiple defs support)", () => {
		expect(
			prepareAnnotations([
				{StartByte: 1, EndByte: 2, URL: "a"},
				{StartByte: 1, EndByte: 3, URL: "a"},
				{StartByte: 2, EndByte: 3, URL: "b"},
				{StartByte: 2, EndByte: 4, Class: "y"},
				{StartByte: 2, EndByte: 3, URL: "c"},
				{StartByte: 2, EndByte: 3, Class: "x"},
				{StartByte: 3, EndByte: 4, URL: "c"},
			])
		).to.eql([
			{StartByte: 1, EndByte: 3, URL: "a"},
			{StartByte: 1, EndByte: 2, URL: "a"},
			{StartByte: 2, EndByte: 4, Class: "y", WantInner: 1},
			{StartByte: 2, EndByte: 3, URLs: ["b", "c"]},
			{StartByte: 2, EndByte: 3, Class: "x", WantInner: 1},
			{StartByte: 3, EndByte: 4, URL: "c"},
		]);
	});
});
