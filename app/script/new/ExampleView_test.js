import shallowRender from "./util/shallowRender";
import expect from "expect.js";

import React from "react";

import ExampleView from "./ExampleView";
import CodeListing from "./CodeListing";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

describe("ExampleView", () => {
	it("should initially render empty and want example", () => {
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<ExampleView defURL="/someURL" examples={{get(defURL, index) { return null; }, getCount(defURL) { return 10; }}} />
			).compare(
				<div className="example">
					<header>
						<div className="pull-right"></div>
						<nav>
							<a className={`fa fa-chevron-circle-left btnNav disabled`}></a>
							<a className={`fa fa-chevron-circle-right btnNav `}></a>
						</nav>
						<i className="fa fa-spinner fa-spin"></i>
					</header>

					<div className="body">
					</div>

					<footer>
						<a target="_blank" href={`/someURL/.examples`} className="pull-right">
							<i className="fa fa-eye" /> View all
						</a>
					</footer>
				</div>
			);
		})).to.eql([new DefActions.WantExample("/someURL", 0)]);
	});

	it("should display available example", () => {
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(
				<ExampleView
					defURL="/someURL"
					examples={{get(defURL, index) { return {Repo: "someRepo", File: "foo.go", StartLine: 3, EndLine: 7, SourceCode: {Lines: [{test: "aLine"}]}}; }, getCount(defURL) { return 10; }}}
					highlightedDef="/otherURL" />
			).compare(
				<div className="example">
					<header>
						<div className="pull-right">someRepo</div>
						<nav>
							<a className={`fa fa-chevron-circle-left btnNav disabled`}></a>
							<a className={`fa fa-chevron-circle-right btnNav `}></a>
						</nav>
						<a>foo.go:3-7</a>
					</header>

					<div className="body">
						<div>
							<CodeListing
								lines={[{test: "aLine"}]}
								selectedDef="/someURL"
								highlightedDef="/otherURL" />
						</div>
					</div>

					<footer>
						<a target="_blank" href="/someURL/.examples" className="pull-right">
							<i className="fa fa-eye" /> View all
						</a>
					</footer>
				</div>
			);
		})).to.eql([new DefActions.WantExample("/someURL", 0)]);
	});
});
