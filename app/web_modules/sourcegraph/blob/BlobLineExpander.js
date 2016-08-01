import * as React from "react";
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
			<tr styleName="line-expander">
				<td styleName="line-expander-cell" onClick={this._onClick}>
					{/* TODO support doing the up/down arrow logic automatically */}
					<div styleName="line-expander-icon">...</div>
				</td>
				{this.state.lineNumbers && <td></td>}
			</tr>
		);
	}
}

export default CSSModules(BlobLineExpander, styles);
