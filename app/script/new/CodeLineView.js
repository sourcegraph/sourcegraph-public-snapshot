import React from "react";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";

// TODO support for tokens with more than one URL
export default class CodeLineView extends Component {
	reconcileState(state, props) {
		// update ownURLs if showing different tokens
		if (state.tokens !== props.tokens) {
			state.tokens = props.tokens;
			state.ownURLs = {};
			state.tokens.forEach((token) => {
				if (token.URL) {
					token.URL.forEach((url) => {
						state.ownURLs[url] = true;
					});
				}
			});
		}

		// filter selectedDef and highlightedDef to improve performance
		state.selectedDef = state.ownURLs[props.selectedDef] ? props.selectedDef : null;
		state.highlightedDef = state.ownURLs[props.highlightedDef] ? props.highlightedDef : null;

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
					{this.state.tokens.map((token, i) => {
						if (!token["URL"]) {
							return <span className={token.Class || ""} key={i}>{token.Label}</span>;
						}

						let cls = `${token.Class || ""} ref`;
						if (token.IsDef) {
							cls += " def";
						}
						let selected = false;
						let highlighted = false;
						token.URL.forEach((url) => {
							selected = selected || url === this.state.selectedDef;
							highlighted = highlighted || url === this.state.highlightedDef;
						});
						if (selected) {
							cls += " highlight-primary";
						}
						if (!selected && highlighted) {
							cls += " highlight-secondary";
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
									if (token.URL.length > 1) {
										Dispatcher.asyncDispatch(new DefActions.SelectMultipleDefs(token.URL, event.clientX, event.clientY)); // dispatch asynchronously so the menu is not immediately closed by click handler on document
										return;
									}
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
	selected: React.PropTypes.bool,
	selectedDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};
