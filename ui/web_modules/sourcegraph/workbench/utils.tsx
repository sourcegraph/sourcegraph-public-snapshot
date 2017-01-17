import * as React from "react";

import URI from "vs/base/common/uri";
import { IEditorInput } from "vs/platform/editor/common/editor";

import { Router, __getRouterForWorkbenchOnly } from "sourcegraph/app/router";

export function getResource(input: IEditorInput): URI {
	if (input["resource"]) {
		return (input as any).resource;
	} else {
		throw new Error("Couldn't find resource.");
	}
}

export const NoopDisposer = { dispose: () => {/* */ } };

export interface PathSpec {
	repo: string;
	rev: string | null;
	path: string;
}

export class RouterContext extends React.Component<{}, {}> {
	static childContextTypes: { [key: string]: React.Validator<any> } = {
		router: React.PropTypes.object.isRequired,
	};

	getChildContext(): { router: Router } {
		const router = __getRouterForWorkbenchOnly();
		router.setRouteLeaveHook = () => {
			throw new Error("Cannot set route leave hook outside React Router hierarchy.");
		};
		router.isActive = () => {
			throw new Error("Cannot access isActive outside React Router hierarchy.");
		};
		return { router };
	}

	render(): JSX.Element {
		return this.props.children as JSX.Element;
	}
}
