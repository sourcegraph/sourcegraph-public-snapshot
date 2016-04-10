// @flow weak

import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import type {State} from "sourcegraph/app/status";
import testOnly from "sourcegraph/util/testOnly";

testOnly();

export function renderedStatus(component): State {
	testOnly();
	let status: State = {error: null};
	let renderer = TestUtils.createRenderer();
	Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {
			status: {
				error(err) {
					if (err && !status.error) status.error = err;
				},
			},
		});
	});
	return status;
}
