// tslint:disable

import * as React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/popover.css";

class Popover extends Component<any, any> {
	constructor(props) {
		super(props);
		this.state = {
			visible: false,
		};
		this._onClick = this._onClick.bind(this);
	}

	componentDidMount() {
		if (global.document) {
			document.addEventListener("click", this._onClick);
		}
	}

	componentWillUnmount() {
		if (global.document) {
			document.removeEventListener("click", this._onClick);
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick(e) {
		let container = this.refs["container"] as HTMLElement;
		let content = this.refs["content"] as HTMLElement;
		if (container && container.contains(e.target)) {
			// Toggle popover visibility when clicking on container or anywhere else
			this.setState({
				visible: !this.state.visible,
			});
		} else if (content && !content.contains(e.target)) {
			// Dismiss popover when clicking on anything else but content.
			this.setState({
				visible: false,
			});
		}
	}

	render(): JSX.Element | null {
		return (
			<div styleName="container" ref="container">
				{this.state.children[0]}
				{this.state.visible &&
					<div ref="content" styleName={`popover_${this.state.left ? "left" : "right"}`} className={this.state.popoverClassName}>
						{this.state.children[1]}
					</div>}
			</div>
		);
	}
}

(Popover as any).propTypes = {
	children: (props, propName, componentName) => {
		let v = React.PropTypes.arrayOf(React.PropTypes.element).isRequired(props, propName, componentName);
		if (v) return v;
		if (props.children.length !== 2) {
			return new Error("Popover must be constructed with exactly two children.");
			// TODO(chexee): make this accomodate multiple lengths!
		}
		return null;
	},
	left: React.PropTypes.bool, // position popover content to the left (default: right)
	popoverClassName: React.PropTypes.string,
};

export default CSSModules(Popover, styles);
