// tslint:disable

import * as React from "react";

import Component from "sourcegraph/Component";
import * as styles from "sourcegraph/blob/styles/Blob.css";

export type Range = [number, number];

type Props = {
	direction: any,
	expandRange: any,
	onExpand: any,
}

class BlobLineExpander extends Component<Props, any> {
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
			<tr className={styles.line_expander}>
				<td className={styles.line_expander_cell} onClick={this._onClick}>
					{/* TODO support doing the up/down arrow logic automatically */}
					<div className={styles.line_expander_icon}>...</div>
				</td>
				{this.state.lineNumbers && <td></td>}
			</tr>
		);
	}
}

export default BlobLineExpander;
