import TestUtils from "react-addons-test-utils";
import expect from "expect.js";

export default function(instance) {
	let renderer = TestUtils.createRenderer();
	renderer.render(instance);
	return expect(renderer.getRenderOutput());
}
