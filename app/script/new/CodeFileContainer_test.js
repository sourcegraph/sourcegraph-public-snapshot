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

describe("CodeFileContainer", () => {
	it("should handle unavailable file", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", null));
		autotest(testdataUnavailableFile, `${__dirname}/testdata/CodeFileContainer-unavailableFile.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" />
		);
	});

	it("should handle available file and unavailable definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", null));
		autotest(testdataUnavailableDefinition, `${__dirname}/testdata/CodeFileContainer-unavailableDefinition.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
		);
	});

	it("should handle available file and available definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("otherDef"));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", {test: "defData"}));
		Dispatcher.directDispatch(DefStore, new DefActions.ExampleFetched("foo", {test: "exampleData"}));
		Dispatcher.directDispatch(DefStore, new DefActions.DiscussionsFetched("someDef", [{test: "discussionData"}]));
		autotest(testdataAvailableDefinition, `${__dirname}/testdata/CodeFileContainer-availableDefinition.json`,
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
		);
	});
});
