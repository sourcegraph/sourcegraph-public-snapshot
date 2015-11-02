import autotest from "./util/autotest";

import React from "react";

import CodeListing from "./CodeListing";

import testdataLines from "./testdata/CodeListing-lines.json";
import testdataNoLineNumbers from "./testdata/CodeListing-noLineNumbers.json";

describe("CodeListing", () => {
	it("should render lines", () => {
		autotest(testdataLines, `${__dirname}/testdata/CodeListing-lines.json`,
			<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}]} lineNumbers={true} selectedDef="someDef" highlightedDef="otherDef" />
		);
	});

	it("should not render line numbers by default", () => {
		autotest(testdataNoLineNumbers, `${__dirname}/testdata/CodeListing-noLineNumbers.json`,
			<CodeListing lines={[{}]} selectedDef={null} highlightedDef={null} />
		);
	});
});
