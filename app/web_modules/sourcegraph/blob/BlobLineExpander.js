import React from "react";

import Component from "sourcegraph/Component";

class BlobLineExpander extends Component {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick(e) {
		this.state.onExpand(this.state.expandRange);
	}

	render() {
		return (
			<tr className="line-expander" onClick={this._onClick}>
				<td className="line-expander-toggle">
				</td>
				<td className="line-content">
					<i className="fa fa-angle-double-up"></i><br/>
					<i className="fa fa-angle-double-down"></i>
				</td>
			</tr>
		);
	}
}

export default BlobLineExpander;
