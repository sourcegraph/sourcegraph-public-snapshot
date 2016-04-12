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

export function render(component: any): RenderResult {
	testOnly();
	let result: RenderResult = {
		actions: [],
		element: null,
	};
	let renderer = TestUtils.createRenderer();
	result.actions = Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {status: {error: () => null}});
	});
	result.element = renderer.getRenderOutput();
	return result;
}
