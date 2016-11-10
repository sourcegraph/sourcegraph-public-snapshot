// tslint:disable: typedef ordered-imports

import * as React from "react";

// withTree wraps Component and passes it a "path" property specified
// in the URL route.
//
// If the path refers to a file, a redirect occurs. (TODO: not yet implemented.)
export function withTree(Component) {
	interface Props {
		repo: string;
		rev?: string;
		params: any;
	}

	type State = any;

	class WithTree extends React.Component<Props, State> {
		render(): JSX.Element | null {
			let path;
			if (this.props.params.splat instanceof Array) {
				path = this.props.params.splat[1];
			}
			if (!path) {
				path = "/";
			}
			return <Component {...Object.assign({}, this.props, {path: path})} />;
		}
	}

	return WithTree;
}
