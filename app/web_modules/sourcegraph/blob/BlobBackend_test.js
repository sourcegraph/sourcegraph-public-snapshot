import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import BlobBackend from "sourcegraph/blob/BlobBackend";
import prepareAnnotations from "sourcegraph/blob/prepareAnnotations";
import * as BlobActions from "sourcegraph/blob/BlobActions";

describe("BlobBackend", () => {
	it("should handle WantFile", () => {
		BlobBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.api/repos/aRepo@aRev/.tree/aTree?ContentsAsString=true");
			callback(null, null, "someFile");
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "aRev", "aTree"));
		})).to.eql([new BlobActions.FileFetched("aRepo", "aRev", "aTree", "someFile")]);
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
