import expect from "expect.js";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import {BlobBackend} from "sourcegraph/blob/BlobBackend";
import {prepareAnnotations} from "sourcegraph/blob/prepareAnnotations";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {immediateSyncPromise} from "sourcegraph/util/testutil/immediateSyncPromise";

describe("BlobBackend", () => {
	it("should handle WantFile", () => {
		BlobBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/repos/aRepo@aCommitID/-/tree/aPath?ContentsAsString=true&NoSrclibAnns=true");
			return immediateSyncPromise({
				status: 200,
				json: () => "someFile",
			});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "aCommitID", "aPath"));
		})).to.eql([
			new RepoActions.RepoCloning("aRepo", false),
			new BlobActions.FileFetched("aRepo", "aCommitID", "aPath", "someFile" as any),
		]);
	});
	it("should handle WantFile with IncludedAnnotations", () => {
		BlobBackend.fetch = function(url: string, init: RequestInit): Promise<Response> {
			expect(url).to.be("/.api/repos/aRepo@c/-/tree/aPath?ContentsAsString=true&NoSrclibAnns=true");
			return immediateSyncPromise({
				status: 200,
				json: () => ({CommitID: "c", IncludedAnnotations: {Annotations: []}}),
			});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BlobBackend.__onDispatch(new BlobActions.WantFile("aRepo", "c", "aPath"));
		})).to.eql([
			new RepoActions.RepoCloning("aRepo", false),
			new BlobActions.AnnotationsFetched("aRepo", "c", "aPath", 0, 0, {Annotations: []} as any),
			new BlobActions.FileFetched("aRepo", "c", "aPath", {CommitID: "c"} as any),
		]);
	});
});

describe("prepareAnnotations", () => {
	it("should duplicate & set WantInner on syntax highlighting annotations", () => {
		expect(
			prepareAnnotations([
				{StartByte: 10, EndByte: 15, Class: "x", WantInner: 1},
				{StartByte: 20, EndByte: 25, Class: "y", WantInner: 1},
				{StartByte: 30, EndByte: 35, Class: "z", WantInner: 1},
			])
		).to.eql(
			[
				{StartByte: 10, EndByte: 15, Class: "x", WantInner: 0},
				{StartByte: 10, EndByte: 15, Class: "x", WantInner: 1},
				{StartByte: 20, EndByte: 25, Class: "y", WantInner: 0},
				{StartByte: 20, EndByte: 25, Class: "y", WantInner: 1},
				{StartByte: 30, EndByte: 35, Class: "z", WantInner: 0},
				{StartByte: 30, EndByte: 35, Class: "z", WantInner: 1},
			]
		);
	});
	it("should handle zero annotations", () => {
		expect(
			prepareAnnotations([])
		).to.eql([]);
	});
});
