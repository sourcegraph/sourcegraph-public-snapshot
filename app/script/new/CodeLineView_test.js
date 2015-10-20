import shallowRender from "./util/shallowRender";
import expect from "expect.js";

import React from "react";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import CodeLineView from "./CodeLineView";

describe("CodeLineView", () => {
	// TODO test selectedDef and highlightedDef as soon as https://github.com/facebook/react/issues/4461 is resolved
	it("should render tokens", () => {
		shallowRender(
			<CodeLineView lineNumber={42} tokens={[
				{Label: "foo"},
				{Label: "bar", Class: "b"},
				{Label: "baz", Class: "c"},
				{Label: "ref", Class: "d", URL: ["someURL"]},
				{Label: "def", Class: "e", URL: ["otherURL"], IsDef: true},
			]} />
		).compare(
			<tr className="line">
				<td className="line-number" data-line={42}></td>
				<td className="line-content">
					<span className={""} key={0}>foo</span>
					<span className={"b"} key={1}>bar</span>
					<span className={"c"} key={2}>baz</span>
					<a href="someURL" className={"d ref"} key={3}>ref</a>
					<a href="otherURL" className={"e ref def"} key={4}>def</a>
				</td>
			</tr>
		);
	});

	it("should select definition on click", () => {
		let defaultPrevented = false;
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(<CodeLineView tokens={[{URL: ["someURL"]}]} />).querySelector("a").props.onClick({preventDefault() { defaultPrevented = true; }});
		})).to.eql([new DefActions.SelectDef("someURL")]);
		expect(defaultPrevented).to.be(true);
	});

	it("should highlight definition on mouse-over", () => {
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(<CodeLineView tokens={[{URL: ["someURL"]}]} />).querySelector("a").props.onMouseOver();
		})).to.eql([new DefActions.HighlightDef("someURL")]);
	});

	it("should remove definition highlight on mouse-out", () => {
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(<CodeLineView tokens={[{URL: ["someURL"]}]} />).querySelector("a").props.onMouseOut();
		})).to.eql([new DefActions.HighlightDef(null)]);
	});
});
