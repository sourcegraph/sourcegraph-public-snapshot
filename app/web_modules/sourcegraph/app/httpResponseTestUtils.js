// @flow weak

import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";

export function renderedHTTPStatusCode(component): ?number {
	let statusCode = null;
	let renderer = TestUtils.createRenderer();
	Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {
			httpResponse: {
				setStatusCode(code) {
					statusCode = code;
				},
			},
		});
	});
	return statusCode;
}
