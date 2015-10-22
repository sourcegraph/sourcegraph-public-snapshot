import shallowRender from "./util/shallowRender";

import React from "react";

import CodeListing from "./CodeListing";
import CodeLineView from "./CodeLineView";

describe("CodeListing", () => {
	it("should render lines", () => {
		shallowRender(
			<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}]} lineNumbers={true} selectedDef="someDef" highlightedDef="otherDef" />
		).compare(
			<table className="line-numbered-code">
				<tbody>
					<CodeLineView lineNumber={1} tokens={["foo"]} selectedDef="someDef" highlightedDef="otherDef" key={0} />
					<CodeLineView lineNumber={2} tokens={[]} selectedDef="someDef" highlightedDef="otherDef" key={1} />
					<CodeLineView lineNumber={3} tokens={["bar"]} selectedDef="someDef" highlightedDef="otherDef" key={2} />
				</tbody>
			</table>
		);
	});

	it("should not render line numbers by default", () => {
		shallowRender(
			<CodeListing lines={[{}]} />
		).compare(
			<table className="line-numbered-code">
				<tbody>
					<CodeLineView lineNumber={null} tokens={[]} key={0} />
				</tbody>
			</table>
		);
	});
});
