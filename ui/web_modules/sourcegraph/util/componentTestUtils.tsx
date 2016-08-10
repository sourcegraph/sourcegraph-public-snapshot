// tslint:disable: typedef ordered-imports

import * as TestUtils from "react-addons-test-utils";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {testOnly} from "sourcegraph/util/testOnly";
import * as ReactDOMServer from "react-dom/server";

testOnly();

export function render(component, context?: {[key: string]: any}) {
	testOnly();
	let renderer = TestUtils.createRenderer();
	Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, Object.assign({}, context, {
			status: {error() { /* empty */ }},
		}));
	});
	return renderer.getRenderOutput();
}

export function renderToString(component, context?: {[key: string]: any}) {
	testOnly();
	return ReactDOMServer.renderToStaticMarkup(render(component, context));
}
