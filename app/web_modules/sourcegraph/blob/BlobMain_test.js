// @flow weak

import autotest from "sourcegraph/util/autotest";

import React from "react";

import BlobMain from "sourcegraph/blob/BlobMain";
import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";

import testdataUnavailableFile from "sourcegraph/blob/testdata/BlobMain-unavailableFile.json";
import testdataUnavailableDefinition from "sourcegraph/blob/testdata/BlobMain-unavailableDefinition.json";
import testdataAvailableDefinition from "sourcegraph/blob/testdata/BlobMain-availableDefinition.json";
import testdataFileFromDef from "sourcegraph/blob/testdata/BlobMain-fileFromDef.json";

describe("BlobMain", () => {
	let exampleFile = {
		ContentsString: "hello\nworld",
	};

	it("should handle unavailable file", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("aRepo", "aRev", "aPath", null));
		autotest(testdataUnavailableFile, `${__dirname}/testdata/BlobMain-unavailableFile.json`,
			<BlobMain repo="aRepo" rev="aRev" path="aPath" />
		);
	});

	it("should handle available file and unavailable definition", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("aRepo", "aRev", "aPath", exampleFile));
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {Error: true}));
		autotest(testdataUnavailableDefinition, `${__dirname}/testdata/BlobMain-unavailableDefinition.json`,
			<BlobMain repo="aRepo" rev="aRev" path="aPath" activeDef="/r@v/-/def/d" />
		);
	});

	it("should handle available file and available definition", () => {
		BlobStore.directDispatch(new BlobActions.FileFetched("r", "v", "p", exampleFile));
		DefStore.directDispatch(new DefActions.HighlightDef("/r@v/-/def/d"));
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {File: {Path: "p"}}));
		DefStore.directDispatch(new DefActions.RefsFetched("r", "v", "d", null, [{test: "exampleData"}]));
		autotest(testdataAvailableDefinition, `${__dirname}/testdata/BlobMain-availableDefinition.json`,
			<BlobMain repo="r" rev="v" path="p" activeDef="/r@v/-/def/d" />
		);
	});

	it("should get filename from definition", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {File: "p", Kind: "someKind"}));
		autotest(testdataFileFromDef, `${__dirname}/testdata/BlobMain-fileFromDef.json`,
			<BlobMain repo="aRepo" rev="aRev" path="p" activeDef="someDef" />
		);
	});
});
