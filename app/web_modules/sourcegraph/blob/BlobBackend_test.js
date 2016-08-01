import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import BlobBackend from "sourcegraph/blob/BlobBackend";
import prepareAnnotations from "sourcegraph/blob/prepareAnnotations";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import immediateSyncPromise from "sourcegraph/util/immediateSyncPromise";


describe("BlobBackend", () => {
	it("should handle WantFile", () => {
		BlobBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos/aRepo@aCommitID/-/tree/aPath?ContentsAsString=true");
			return immediateSyncPromise({
				status: 200,
				json: () => "someFile",
			});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "aCommitID", "aPath"));
		})).to.eql([
			new RepoActions.RepoCloning("aRepo", false),
			new BlobActions.FileFetched("aRepo", "aCommitID", "aPath", "someFile"),
		]);
	});
	it("should handle WantFile with IncludedAnnotations", () => {
		BlobBackend.fetch = function(url, options) {
			expect(url).to.be("/.api/repos/aRepo@c/-/tree/aPath?ContentsAsString=true");
			return immediateSyncPromise({
				status: 200,
				json: () => ({CommitID: "c", IncludedAnnotations: {Annotations: []}}),
			});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "c", "aPath"));
		})).to.eql([
			new RepoActions.RepoCloning("aRepo", false),
			new BlobActions.AnnotationsFetched("aRepo", "c", "aPath", 0, 0, {Annotations: []}),
			new BlobActions.FileFetched("aRepo", "c", "aPath", {CommitID: "c"}),
		]);
	});
});

describe("prepareAnnotations", () => {
	it("should set WantInner on syntax highlighting annotations", () => {
		expect(
			prepareAnnotations([
				{StartByte: 10, EndByte: 15, Class: "x"},
				{StartByte: 20, EndByte: 25, URL: "y"},
				{StartByte: 30, EndByte: 35, Class: "z"},
			])
		).to.eql(
			[
				{StartByte: 10, EndByte: 15, Class: "x", WantInner: 1},
				{StartByte: 20, EndByte: 25, URL: "y"},
				{StartByte: 30, EndByte: 35, Class: "z", WantInner: 1},
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
	it("should handle zero annotations", () => {
		expect(
			prepareAnnotations([])
		).to.eql([]);
	});
});
