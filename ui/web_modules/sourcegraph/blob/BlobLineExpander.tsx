// tslint:disable: typedef ordered-imports

import * as React from "react";

import {Component} from "sourcegraph/Component";
import * as styles from "sourcegraph/blob/styles/Blob.css";

export type Range = [number, number];

interface Props {
	direction: any;
	expandRange: any;
	onExpand: any;
}

export class BlobLineExpander extends Component<Props, any> {
	constructor(props: Props) {
		super(props);
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props: Props) {
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
