import React from "react";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

// TODO support for tokens with more than one URL
export default class CodeLineView extends Component {
	reconcileState(state, props) {
		// update ownURLs if showing different tokens
		if (state.tokens !== props.tokens) {
			state.tokens = props.tokens;
			state.ownURLs = {};
			state.tokens.forEach((token) => {
				if (token["URL"]) {
					state.ownURLs[token.URL[0]] = true;
				}
			});
		}

		// filter selectedDef and highlightedDef to improve performance
		state.selectedDef = state.ownURLs[props.selectedDef] ? props.selectedDef : null;
		state.highlightedDef = state.ownURLs[props.highlightedDef] ? props.highlightedDef : null;

		state.lineNumber = props.lineNumber || null;
	}

	render() {
		return (
			<tr className="line">
				{this.state.lineNumber && <td className="line-number" data-line={this.state.lineNumber}></td>}
				<td className="line-content">
					{this.state.tokens.map((token, i) => {
						if (!token["URL"]) {
							return <span className={token.Class || ""} key={i}>{token.Label}</span>;
						}

						let cls = `${token.Class || ""} ref`;
						if (token.IsDef) {
							cls += " def";
						}
						switch (token.URL[0]) {
						case this.state.selectedDef:
							cls += " highlight-primary";
							break;
						case this.state.highlightedDef:
							cls += " highlight-secondary";
							break;
						}
						return (
							<a
								className={cls}
								href={token.URL[0]}
								onMouseOver={() => {
									Dispatcher.dispatch(new DefActions.HighlightDef(token.URL[0]));
								}}
								onMouseOut={() => {
									Dispatcher.dispatch(new DefActions.HighlightDef(null));
								}}
								onClick={(event) => {
									event.preventDefault();
									Dispatcher.dispatch(new DefActions.SelectDef(token.URL[0]));
								}}
								key={i}>
								{token.Label}
							</a>
						);
					})}
					{this.state.tokens.length === 0 && <span>&nbsp;</span>}
				</td>
			</tr>
		);
	}
}

CodeLineView.propTypes = {
	lineNumber: React.PropTypes.number,
	tokens: React.PropTypes.array,
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};
