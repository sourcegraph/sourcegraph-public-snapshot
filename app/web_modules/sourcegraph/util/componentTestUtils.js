// @flow weak

import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import testOnly from "sourcegraph/util/testOnly";
import ReactDOMServer from "react-dom/server";

testOnly();

export function render(component) {
	testOnly();
	let renderer = TestUtils.createRenderer();
	Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {
			status: {error() {}},
		});
	});
	return renderer.getRenderOutput();
}

export function renderToString(component) {
	testOnly();
	return ReactDOMServer.renderToStaticMarkup(render(component));
}
