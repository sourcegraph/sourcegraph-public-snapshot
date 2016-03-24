import autotest from "sourcegraph/util/autotest";

import React from "react";

import BlobContainer from "sourcegraph/blob/BlobContainer";
import BlobStore from "sourcegraph/blob/BlobStore";
import DefStore from "sourcegraph/def/DefStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";

import testdataUnavailableFile from "sourcegraph/blob/testdata/BlobContainer-unavailableFile.json";
import testdataUnavailableDefinition from "sourcegraph/blob/testdata/BlobContainer-unavailableDefinition.json";
import testdataAvailableDefinition from "sourcegraph/blob/testdata/BlobContainer-availableDefinition.json";
import testdataFileFromDef from "sourcegraph/blob/testdata/BlobContainer-fileFromDef.json";
import testdataDefOptions from "sourcegraph/blob/testdata/BlobContainer-defOptions.json";

describe("BlobContainer", () => {
	let exampleFile = {
		ContentsString: "hello\nworld",
	};

	it("should handle unavailable file", () => {
		Dispatcher.directDispatch(BlobStore, new BlobActions.FileFetched("aRepo", "aRev", "aPath", null));
		autotest(testdataUnavailableFile, `${__dirname}/testdata/BlobContainer-unavailableFile.json`,
			<BlobContainer repo="aRepo" rev="aRev" path="aPath" />
		);
	});

	it("should handle available file and unavailable definition", () => {
		Dispatcher.directDispatch(BlobStore, new BlobActions.FileFetched("aRepo", "aRev", "aPath", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", null));
		autotest(testdataUnavailableDefinition, `${__dirname}/testdata/BlobContainer-unavailableDefinition.json`,
			<BlobContainer repo="aRepo" rev="aRev" path="aPath" activeDef="someDef" />
		);
	});

	it("should handle available file and available definition", () => {
		Dispatcher.directDispatch(BlobStore, new BlobActions.FileFetched("aRepo", "aRev", "aPath", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("otherDef"));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("otherDef", {File: {Path: "aPath"}}));
		Dispatcher.directDispatch(DefStore, new DefActions.RefsFetched("foo", null, [{test: "exampleData"}]));
		autotest(testdataAvailableDefinition, `${__dirname}/testdata/BlobContainer-availableDefinition.json`,
			<BlobContainer repo="aRepo" rev="aRev" path="aPath" activeDef="someDef" />
		);
	});

	it("should get filename from definition", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", {File: "somePath", Kind: "someKind"}));
		autotest(testdataFileFromDef, `${__dirname}/testdata/BlobContainer-fileFromDef.json`,
			<BlobContainer repo="aRepo" rev="aRev" activeDef="someDef" />
		);
	});

	it("should render def options menu", () => {
		Dispatcher.directDispatch(BlobStore, new BlobActions.FileFetched("aRepo", "aRev", "aPath", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.SelectMultipleDefs(["firstDef", "secondDef"], 10, 20));
		autotest(testdataDefOptions, `${__dirname}/testdata/BlobContainer-defOptions.json`,
			<BlobContainer repo="aRepo" rev="aRev" path="aPath" />
		);
	});
});
