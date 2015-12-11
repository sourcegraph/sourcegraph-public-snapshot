import autotest from "sourcegraph/util/autotest";

import React from "react";

import CodeListing from "sourcegraph/code/CodeListing";

import testdataLines from "sourcegraph/code/testdata/CodeListing-lines.json";
import testdataNoLineNumbers from "sourcegraph/code/testdata/CodeListing-noLineNumbers.json";
import testdataSelectionForm from "sourcegraph/code/testdata/CodeListing-selectionForm.json";

describe("CodeListing", () => {
	it("should render lines", () => {
		autotest(testdataLines, `${__dirname}/testdata/CodeListing-lines.json`,
			<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}, {}]} lineNumbers={true} startLine={1} endLine={2} selectedDef="someDef" highlightedDef="otherDef" />
		);
	});

	it("should not render line numbers by default", () => {
		autotest(testdataNoLineNumbers, `${__dirname}/testdata/CodeListing-noLineNumbers.json`,
			<CodeListing lines={[{}]} selectedDef={null} highlightedDef={null} />
		);
	});

	it("should render line selection form", () => {
		autotest(testdataSelectionForm, `${__dirname}/testdata/CodeListing-selectionForm.json`,
			<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}, {}]} lineNumbers={true} startLine={1} endLine={2} selectedDef="someDef" highlightedDef="otherDef" onLineButtonClick={function() {}}/>
		);
	});
});
