// @flow

import React from "react";
import CSSModules from "react-css-modules";

import Component from "sourcegraph/Component";
import styles from "sourcegraph/blob/styles/Blob.css";

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
			<tr styleName="line-expander" onClick={this._onClick}>
				<td>
				</td>
				<td>
					{/* TODO support doing the up/down arrow logic automatically */}
					{this.state.direction !== "down" && <span className="line-expander-icon">&#8679;</span>}
					{this.state.direction !== "up" && <span className="line-expander-icon">&#8681;</span>}
				</td>
			</tr>
		);
	}
}

export default CSSModules(BlobLineExpander, styles);
