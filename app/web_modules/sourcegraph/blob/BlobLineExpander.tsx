// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";

import Component from "sourcegraph/Component";
import * as styles from "sourcegraph/blob/styles/Blob.css";

export type Range = [number, number];

class BlobLineExpander extends Component<any, any> {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick(e: Event) {
		this.state.onExpand(this.state.expandRange);
	}

	render(): JSX.Element | null {
		return (
			<tr styleName="line_expander">
				<td styleName="line_expander_cell" onClick={this._onClick}>
					{/* TODO support doing the up/down arrow logic automatically */}
					<div styleName="line_expander_icon">...</div>
				</td>
				{this.state.lineNumbers && <td></td>}
			</tr>
		);
	}
}

export default CSSModules(BlobLineExpander, styles);
