import TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import testOnly from "sourcegraph/util/testOnly";
import type {Element} from "react";

testOnly();

export type RenderResult = {
	actions: Array<Object>;
	element: ?Element<any>;
};

export function render(component: any, context: ?Object): RenderResult {
	testOnly();
	let result: RenderResult = {
		actions: [],
		element: null,
	};
	let renderer = TestUtils.createRenderer();
	result.actions = Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, {
			router: {},
			...context,
		});
	});
	result.element = renderer.getRenderOutput();
	return result;
}
