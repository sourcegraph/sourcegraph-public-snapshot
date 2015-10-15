import shallowRender from "./shallowRender";
import expect from "expect.js";

import React from "react";

import CodeFileContainer from "./CodeFileContainer";
import CodeListing from "./CodeListing";
import CodeStore from "./CodeStore";
import * as CodeActions from "./CodeActions";
import Dispatcher from "./Dispatcher";

describe("CodeFileContainer", () => {
	it("should handle unavailable file", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", undefined));
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" />
			).compare(
				null
			);
		})).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree")]);
	});

	it("should handle available file", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		Dispatcher.directDispatch(CodeStore, new CodeActions.HighlightDef("otherDef"));
		shallowRender(
			<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
		).compare(
			<CodeListing lines={["someLine"]} selectedDef="someDef" highlightedDef="otherDef" />
		);
	});
});
