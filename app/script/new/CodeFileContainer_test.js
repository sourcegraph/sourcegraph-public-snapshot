import shallowRender from "./util/shallowRender";
import expect from "expect.js";

import React from "react";

import CodeFileContainer from "./CodeFileContainer";
import CodeListing from "./CodeListing";
import DefPopup from "./DefPopup";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import Dispatcher from "./Dispatcher";

describe("CodeFileContainer", () => {
	it("should handle unavailable file", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", null));
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" />
			).compare(
				null
			);
		})).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree")]);
	});

	it("should handle available file and unavailable definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		Dispatcher.directDispatch(DefStore, new DefActions.HighlightDef("otherDef"));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", null));
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
			).compare(
				<div>
					<CodeListing lines={["someLine"]} selectedDef="someDef" highlightedDef="otherDef" />
				</div>
			);
		})).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree"), new DefActions.WantDef("someDef")]);
	});

	it("should handle available file and available definition", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someDef", {test: "defData"}));
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<CodeFileContainer repo="aRepo" rev="aRev" tree="aTree" selectedDef="someDef" />
			).compare(
				<div>
					<CodeListing lines={["someLine"]} selectedDef="someDef" />
					<DefPopup def={{test: "defData"}} />
				</div>
			);
		})).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree"), new DefActions.WantDef("someDef")]);
	});
});
