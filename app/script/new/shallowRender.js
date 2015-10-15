import expect from "expect.js";
import TestUtils from "react-addons-test-utils";

class ElementWrapper {
	constructor(element) {
		this.element = element;
	}

	get props() {
		return this.element.props;
	}

	querySelector(selector) {
		// TODO implement more selectors
		if (this.element.type === selector) {
			return this;
		}
		let children = this.element.props && this.element.props.children;
		if (children === undefined) {
			return null;
		}
		if (children.constructor !== Array) {
			return new ElementWrapper(children).querySelector(selector);
		}
		for (let i = 0; i < children.length; i++) {
			let res = new ElementWrapper(children[i]).querySelector(selector);
			if (res !== null) {
				return res;
			}
		}
		return null;
	}

	compare(expected) {
		if (expected === null || expected.constructor === String) {
			expect(this.element).to.be(expected);
			return;
		}

		expect(this.element.constructor).to.be(expected.constructor);
		expect(this.element.type).to.be(expected.type);

		if (expected.props === undefined) {
			expect(this.element.props).to.be(undefined);
			return;
		}
		Object.keys(expected.props).forEach((key) => {
			if (key === "children") { return; }
			expect(this.element.props[key]).to.eql(expected.props[key]);
		});

		if (expected.props.children === undefined) {
			expect(this.element.props.children).to.be(undefined);
			return;
		}
		if (expected.props.children.constructor !== Array) {
			expect(this.element.props.children.constructor).to.not.be(Array);
			new ElementWrapper(this.element.props.children).compare(expected.props.children);
			return;
		}
		expect(this.element.props.children.constructor).to.be(Array);
		expect(this.element.props.children.length).to.be(expected.props.children.length);
		for (let i = 0; i < this.element.props.children.length; i++) {
			new ElementWrapper(this.element.props.children[i]).compare(expected.props.children[i]);
		}
	}
}

export default function(instance, expected) {
	let renderer = TestUtils.createRenderer();
	renderer.render(instance);
	return new ElementWrapper(renderer.getRenderOutput());
}
