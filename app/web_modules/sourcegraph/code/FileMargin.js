import React from "react";
import Component from "sourcegraph/Component";

export default class FileMargin extends Component {
	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
	}

	reconcileState(state, props) {
		state.getOffsetTopForByte = props.getOffsetTopForByte;
		state.children = props.children && state.getOffsetTopForByte ?
			React.Children.map(props.children, (child) => (
				{top: state.getOffsetTopForByte(child.props.byte), component: child}
			)) : null;
	}

	render() {
		return (
			<div className="sidebar file-sidebar" style={{position: "relative"}}>
				{this.state.children && this.state.children.map((child, i) => (
					<div key={i} style={{position: "absolute", top: `${child.top}px`}}>
						{child.component}
					</div>
				))}
			</div>
		);
	}
}
FileMargin.propTypes = {
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
	]),
	getOffsetTopForByte: React.PropTypes.func,
};
