import sandbox from "../testSandbox";
import expect from "expect.js";

import React from "react";

import CodeFileController from "./CodeFileController";
import CodeListing from "./CodeListing";
import * as CodeActions from "./CodeActions";
import Dispatcher from "./Dispatcher";

describe("CodeFileController", () => {
	it("should handle unavailable file", () => {
		Dispatcher.dispatch(new CodeActions.FileFetched("aRepo", "aRev", "aTree", undefined));
		expect(Dispatcher.catchDispatched(() => {
			sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(null);
		})).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree")]);
	});
	
	it("should handle available file", () => {
		Dispatcher.dispatch(new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(
			<CodeListing lines={["someLine"]} />
		);
	});
});
