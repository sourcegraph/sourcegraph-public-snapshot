// @flow

import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import testOnly from "sourcegraph/util/testOnly";
import type {Element} from "react";

testOnly();

export type RenderResult = {
	actions: Array<Object>;
	element: ?Element;
};

export function render(component: any, context: ?Object): RenderResult {
	testOnly();
	let result: RenderResult = {
		actions: [],
		element: null,
	};
	let renderer = TestUtils.createRenderer();
	result.actions = Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {...context, status: {error: () => null}});
	});
	result.element = renderer.getRenderOutput();
	return result;
}
