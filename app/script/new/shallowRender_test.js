import shallowRender from "./shallowRender";

import React from "react";
import expect from "expect.js";

class TestComponent extends React.Component {
	render() {
		return (
			<div><span className="abc">test</span></div>
		);
	}
}

describe("shallowRender(...).compare", () => {
	it("should succeed if equal", () => {
		shallowRender(
			<TestComponent />
		).compare(
			<div><span className="abc">test</span></div>
		);
	});

	it("should compare text", () => {
		expect(() => {
			shallowRender(
				<TestComponent />
			).compare(
				<div><span className="abc">X</span></div>
			);
		}).to.throwException();
	});

	it("should compare tags", () => {
		expect(() => {
			shallowRender(
				<TestComponent />
			).compare(
				<div><a className="abc">test</a></div>
			);
		}).to.throwException();
	});

	it("should compare property", () => {
		expect(() => {
			shallowRender(
				<TestComponent />
			).compare(
				<div><span className="X">test</span></div>
			);
		}).to.throwException();
	});

	it("should ignore properties which are not expected", () => {
		shallowRender(
			<TestComponent />
		).compare(
			<div><span>test</span></div>
		);
	});
});
