// tslint:disable

import * as React from "react";

// withTree wraps Component and passes it a "path" property specified
// in the URL route.
//
// If the path refers to a file, a redirect occurs. (TODO: not yet implemented.)
export default function withTree(Component) {
	class WithTree extends React.Component<any, any> {
		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string,
			commitID: React.PropTypes.string,
			params: React.PropTypes.object.isRequired,
		};

		render(): JSX.Element | null {
			let path;
			if (this.props.params.splat instanceof Array) path = this.props.params.splat[1];
			if (!path) path = "/";
			return <Component {...Object.assign({}, this.props, {path: path})} />;
		}
	}

	return WithTree;
}
