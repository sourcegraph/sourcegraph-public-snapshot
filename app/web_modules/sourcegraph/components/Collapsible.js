import * as React from "react";

import Component from "sourcegraph/Component";

class Collapsible extends Component {
	constructor(props) {
		super(props);
		this.state = {
			shown: !props.collapsed,
		};
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick() {
		const isShown = !this.state.shown;
		this.setState({
			shown: isShown,
		}, () => this.state.onToggle && this.state.onToggle(isShown));
	}

	render() {
		return (
			<div>
				<div onClick={this._onClick} style={{cursor: "pointer"}}>{this.state.children[0]}</div>
				{this.state.shown && this.state.children[1]}
			</div>
		);
	}
}

Collapsible.propTypes = {
	children: (props, propName, componentName) => {
		let v = React.PropTypes.arrayOf(React.PropTypes.element).isRequired(props, propName, componentName);
		if (v) return v;
		if (props.children.length !== 2) {
			return new Error("Collapsible must be constructed with exactly two children.");
		}
		return null;
	},
	collapsed: React.PropTypes.bool,
	onToggle: React.PropTypes.func,
};

export default Collapsible;
