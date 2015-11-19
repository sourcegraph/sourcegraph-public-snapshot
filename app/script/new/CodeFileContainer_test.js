import autotest from "./util/autotest";

import React from "react";

import CodeFileContainer from "./CodeFileContainer";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import Dispatcher from "./Dispatcher";

import testdataUnavailableFile from "./testdata/CodeFileContainer-unavailableFile.json";
import testdataUnavailableDefinition from "./testdata/CodeFileContainer-unavailableDefinition.json";
import testdataAvailableDefinition from "./testdata/CodeFileContainer-availableDefinition.json";
import testdataFileFromDef from "./testdata/CodeFileContainer-fileFromDef.json";
import testdataDefOptions from "./testdata/CodeFileContainer-defOptions.json";

describe("CodeFileContainer", () => {
	let exampleFile = {
		Entry: {SourceCode: {Lines: ["someLine"]}},
		EntrySpec: {RepoRev: {CommitID: "123abc"}},
	};

	it("should handle unavailable file", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", null));
		autotest(testdataUnavailableFile, `${__dirname}/testdata/CodeFileContainer-unavailableFile.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef={null} />
		);
	});

	it("should handle available file and unavailable definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", null));
		autotest(testdataUnavailableDefinition, `${__dirname}/testdata/CodeFileContainer-unavailableDefinition.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
		);
	});

	it("should handle available file and available definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("otherDef"));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", {Found: true, test: "defData"}));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("otherDef", {Found: true, test: "otherDefData"}));
		Dispatcher.directDispatch(DefStore, new DefActions.ExampleFetched("foo", {test: "exampleData"}));
		autotest(testdataAvailableDefinition, `${__dirname}/testdata/CodeFileContainer-availableDefinition.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
		);
	});

	it("should get filename from definition", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", {Found: true, File: {Path: "somePath"}}));
		autotest(testdataFileFromDef, `${__dirname}/testdata/CodeFileContainer-fileFromDef.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" def="someDef" />
		);
	});

	it("should render def options menu", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", exampleFile));
		Dispatcher.directDispatch(DefStore, new DefActions.SelectMultipleDefs(["firstDef", "secondDef"], 10, 20));
		autotest(testdataDefOptions, `${__dirname}/testdata/CodeFileContainer-defOptions.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef={null} />
		);
	});
});
