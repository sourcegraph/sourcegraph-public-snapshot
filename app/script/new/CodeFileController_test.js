import sandbox from "../testSandbox";
import expect from "expect.js";

import React from "react";

import CodeFileController from "./CodeFileController";
import CodeListing from "./CodeListing";
import * as CodeActions from "./CodeActions";
import CodeStore from "./CodeStore";

describe("CodeFileController", () => {
	it("should handle unavailable file", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", undefined));
		sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(null);
		expect(sandbox.dispatched).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree")]);
	});

	it("should handle available file", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(
			<CodeListing lines={["someLine"]} />
		);
	});
});
