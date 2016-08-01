import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import testOnly from "sourcegraph/util/testOnly";
import ReactDOMServer from "react-dom/server";

testOnly();

export function render(component, context?: {[key: string]: any}) {
	testOnly();
	let renderer = TestUtils.createRenderer();
	Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {
			...context,
			status: {error() {}},
		});
	});
	return renderer.getRenderOutput();
}

export function renderToString(component, context?: {[key: string]: any}) {
	testOnly();
	return ReactDOMServer.renderToStaticMarkup(render(component, context));
}
