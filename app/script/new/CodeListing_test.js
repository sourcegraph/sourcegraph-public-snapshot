import shallowRender from "./shallowRender";

import React from "react";

import CodeListing from "./CodeListing";
import CodeLineView from "./CodeLineView";

describe("CodeListing", () => {
	it("should render lines", () => {
		shallowRender(
			<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}]} highlightedDef="someDef" />
		).compare(
			<div className="code-view-react">
				<table className="line-numbered-code">
					<tbody>
						<CodeLineView lineNumber={1} tokens={["foo"]} highlightedDef="someDef" key={0} />
						<CodeLineView lineNumber={2} tokens={[]} highlightedDef="someDef" key={1} />
						<CodeLineView lineNumber={3} tokens={["bar"]} highlightedDef="someDef" key={2} />
					</tbody>
				</table>
			</div>
		);
	});
});
