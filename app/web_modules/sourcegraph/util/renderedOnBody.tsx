// tslint:disable

import * as React from "react";
import * as ReactDOM from "react-dom";
import CSSPropertyOperations from "react/lib/CSSPropertyOperations";
import shallowCompare from "react/lib/shallowCompare";

// Adapted from https://raw.githubusercontent.com/tajo/react-portal/master/lib/portal.js.

export default function renderedOnBody(Component) {
	class RenderedOnBody extends React.Component<any, any> {
		_node: any;
		_elem: any;

		static propTypes = {
			className: React.PropTypes.string,
			style: React.PropTypes.object,
			children: React.PropTypes.element.isRequired,
		};

		componentDidMount() {
			this._renderOnBody(this.props);
		}

		componentWillReceiveProps(newProps) {
			this._renderOnBody(newProps);
		}

		shouldComponentUpdate(nextProps, nextState) {
			return shallowCompare(this, nextProps, nextState);
		}

		componentWillUnmount() {
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
				if (props.style) {
					CSSPropertyOperations.setValueForStyles(this._node, props.style);
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
