import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as CodeActions from "sourcegraph/code/CodeActions";

class CodeLineView extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.lineNumber = props.lineNumber || null;
		state.selected = Boolean(props.selected);
	}

	render() {
		return (
			<tr className={`line ${this.state.selected ? "main-byte-range" : ""}`}>
				{this.state.lineNumber &&
					<td className="line-number"
						data-line={this.state.lineNumber}
						onClick={(event) => {
							if (event.shiftKey) {
								Dispatcher.dispatch(new CodeActions.SelectRange(this.state.lineNumber));
								return;
							}
							Dispatcher.dispatch(new CodeActions.SelectLine(this.state.lineNumber));
						}}>
					</td>}
				<td className="line-content">
					{this.state.contents}
					{this.state.contents === "" && <span>&nbsp;</span>}
				</td>
			</tr>
		);
	}
}

CodeLineView.propTypes = {
	lineNumber: React.PropTypes.number,
	contents: React.PropTypes.string,
	selected: React.PropTypes.bool,
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};

export default CodeLineView;
