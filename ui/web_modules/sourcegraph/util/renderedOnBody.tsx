// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as ReactDOM from "react-dom";
import shallowCompare from "react/lib/shallowCompare";

// Adapted from https://raw.githubusercontent.com/tajo/react-portal/master/lib/portal.js.

export function renderedOnBody(Component) {
	interface Props {
		className?: string;
		style?: any;

		[key: string]: any;
	}

	type State = any;

	class RenderedOnBody extends React.Component<Props, State> {
		_node: any;
		_elem: any;

		componentDidMount(): void {
			this._renderOnBody(this.props);
		}

		componentWillReceiveProps(newProps) {
			this._renderOnBody(newProps);
		}

		shouldComponentUpdate(nextProps, nextState) {
			return shallowCompare(this, nextProps, nextState);
		}

		componentWillUnmount(): void {
			if (this._node) {
				ReactDOM.unmountComponentAtNode(this._node);
				document.body.removeChild(this._node);
			}
			this._elem = null;
			this._node = null;
		}

		_renderOnBody(props) {
			if (!this._node) {
				this._node = document.createElement("div");
				if (props.className) {
					this._node.className = props.className;
				}
				document.body.appendChild(this._node);
			}
			this._elem = ReactDOM.unstable_renderSubtreeIntoContainer(this, React.cloneElement(<Component {...props} />), this._node);
		}

		render(): JSX.Element | null {
			return null;
		}
	}
	return RenderedOnBody;
}
