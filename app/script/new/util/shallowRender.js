import expect from "expect.js";
import TestUtils from "react-addons-test-utils";
import mockTimeout from "./mockTimeout";

class ElementWrapper {
	constructor(element) {
		this.element = element;
	}

	// props returns the properties of the element.
	get props() {
		return this.element.props;
	}

	// querySelector returns the first child matching the CSS selector.
	querySelector(selector) {
		// TODO implement more selectors
		if (selector.charAt(0) === "." && this.element.props && this.element.props.className && this.element.props.className.split(" ").indexOf(selector.substr(1)) !== -1) {
			return this;
		} else if (this.element.type === selector) {
			return this;
		}
		let children = toChildArray(this.element.props && this.element.props.children);
		for (let i = 0; i < children.length; i++) {
			let res = new ElementWrapper(children[i]).querySelector(selector);
			if (res !== null) {
				return res;
			}
		}
		return null;
	}

	// compare compares this element with expected and raise error on mismatch.
	// Properties which are in the element but not in the expectation are skipped.
	compare(expected) {
		if (expected === null || expected.constructor === String) {
			expect(this.element).to.be(expected);
			return;
		}

		expect(this.element.constructor).to.be(expected.constructor);
		expect(this.element.type).to.be(expected.type);

		if (!expected.props) {
			expect(this.element.props).to.eql(false);
			return;
		}
		Object.keys(expected.props).forEach((key) => {
			if (key === "children" || key === "zIndex") { return; }
			expect(this.element.props[key]).to.eql(expected.props[key]);
		});

		let thisChildren = mergeText(toChildArray(this.element.props.children));
		let expectedChildren = mergeText(toChildArray(expected.props.children));
		expect(thisChildren.length).to.be(expectedChildren.length);
		for (let i = 0; i < thisChildren.length; i++) {
			new ElementWrapper(thisChildren[i]).compare(expectedChildren[i]);
		}
	}
}

function toChildArray(children) {
	if (!children) {
		return [];
	}
	if (children.constructor !== Array) {
		return [children];
	}
	return children.reduce((a, e) => a.concat(toChildArray(e)), []);
}

function mergeText(elements) {
	let merged = [];
	elements.forEach((e) => {
		if (e.constructor === Number) {
			e = String(e);
		}
		let i = merged.length - 1;
		if (e.constructor === String && i !== -1 && merged[i].constructor === String) {
			merged[i] += e;
			return;
		}
		merged.push(e);
	});
	return merged;
}

// Shallow render the given component. Does not use the DOM.
export default function(instance, expected) {
	let renderer = TestUtils.createRenderer();
	mockTimeout(() => {
		renderer.render(instance);
	});
	return new ElementWrapper(renderer.getRenderOutput());
}
