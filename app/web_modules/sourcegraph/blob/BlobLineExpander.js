// @flow

import React from "react";

import Component from "sourcegraph/Component";

export type Range = [number, number];

class BlobLineExpander extends Component {
	constructor(props: BlobLineExpander.props) {
		super(props);
		this._onClick = this._onClick.bind(this);
	}

	state: BlobLineExpander.props;

	props: {
		expandRange: Range;
		onExpand: (range: Range) => any;
	};

	reconcileState(state: BlobLineExpander.state, props: BlobLineExpander.props) {
		Object.assign(state, props);
	}

	_onClick(e: Event) {
		this.state.onExpand(this.state.expandRange);
	}

	render() {
		return (
			<tr className="line-expander" onClick={this._onClick}>
				<td className="line-expander-toggle">
				</td>
				<td className="line-content">
					{/* TODO support doing the up/down arrow logic automatically */}
					{this.state.direction !== "down" && <i className="fa fa-angle-double-up expand-icon"></i>}
					{this.state.direction !== "up" && <i className="fa fa-angle-double-down expand-icon"></i>}
				</td>
			</tr>
		);
	}
}

export default BlobLineExpander;
