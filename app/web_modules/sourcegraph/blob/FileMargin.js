// @flow weak

import React from "react";
import Component from "sourcegraph/Component";

export default class FileMargin extends Component {
	constructor(props) {
		super(props);
		this.state = {extraPadding: 0};
	}

	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
	}

	componentDidUpdate() {
		this._calculateExtraPadding();
	}

	reconcileState(state, props) {
		state.getOffsetTopForByte = props.getOffsetTopForByte;
		state.children = props.children && state.getOffsetTopForByte ?
			React.Children.map(props.children, (child) => (
				{top: state.getOffsetTopForByte(child.props.byte), component: child}
			)) : null;
	}

	// _calculateExtraPadding will update the additional padding needed on the
	// bottom of the sidebar to prevent the window from needing a scrollbar for
	// any overflowing children.
	_calculateExtraPadding() {
		let el = this.refs && this.refs.sidebar;
		if (!el) return;

		for (let child of el.children) {
			let offset = child.offsetHeight + parseInt(child.dataset.offset, 10);
			if (offset > (el.offsetHeight - this.state.extraPadding)) {
				// Only update when this changes to prevent stack overflow.
				if (offset === this.state.extraPadding) return;
				this.setState({extraPadding: offset});
				return;
			}
		}

		if (this.state.extraPadding !== 0) this.setState({extraPadding: 0});
	}

	render() {
		return (
			<div className={this.props.className} ref="sidebar" style={{paddingBottom: this.state.extraPadding}}>
				{this.state.children && this.state.children.map((child, i) => (
					<div key={i} style={{width: "100%", position: "absolute", top: `${child.top}px`}} data-offset={child.top}>
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
	className: React.PropTypes.string,
};
