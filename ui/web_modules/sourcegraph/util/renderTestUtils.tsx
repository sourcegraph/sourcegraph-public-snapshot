// tslint:disable

import * as TestUtils from "react-addons-test-utils";
import Dispatcher from "sourcegraph/Dispatcher";
import testOnly from "sourcegraph/util/testOnly";

testOnly();

export type RenderResult = {
	actions: Array<Object>;
	element: JSX.Element | null;
};

export function render(component: any, context?: any): RenderResult {
	testOnly();
	let result: RenderResult = {
		actions: [],
		element: null,
	};
	let renderer = TestUtils.createRenderer();
	result.actions = Dispatcher.Backends.catchDispatched(() => {
		renderer.render(component, Object.assign({
			router: {},
		}, context));
	});
	result.element = renderer.getRenderOutput();
	return result;
}
