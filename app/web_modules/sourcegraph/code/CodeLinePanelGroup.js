import React from "react";

import Component from "sourcegraph/Component";

const width = 444;
const marginRight = 10;
const marginTop = 3;

class CodeLinePanelGroup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let styles = {
			width: width,
			right: marginRight,
			marginTop: marginTop - 15, // Subtract width of the current line.
		};

		return (
			<div className="line-panel-group" style={styles}>
				{this.state.items.map((item, i) => <span key={i}>{item}</span>)}
			</div>
		);
	}
}

export default CodeLinePanelGroup;
