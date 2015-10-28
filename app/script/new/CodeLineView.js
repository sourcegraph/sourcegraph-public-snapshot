import React from "react";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

// TODO support for tokens with more than one URL
export default class CodeLineView extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
		this.state = this._calculateState(props);
	}

	componentWillReceiveProps(nextProps) {
		this.setState(this._calculateState(nextProps));
	}

	shouldComponentUpdate(nextProps, nextState) {
		return nextState.tokens !== this.state.tokens ||
			nextState.selectedDef !== this.state.selectedDef ||
			nextState.highlightedDef !== this.state.highlightedDef;
	}

	_calculateState(props) {
		let ownURLs = this.state.ownURLs;
		if (this.state.tokens !== props.tokens) {
			ownURLs = {};
			props.tokens.forEach((token) => {
				if (token["URL"]) {
					ownURLs[token.URL[0]] = true;
				}
			});
		}

		// filter selectedDef and highlightedDef to improve performance
		return {
			tokens: props.tokens,
			ownURLs: ownURLs,
			selectedDef: ownURLs[props.selectedDef] ? props.selectedDef : null,
			highlightedDef: ownURLs[props.highlightedDef] ? props.highlightedDef : null,
		};
	}

	render() {
		return (
			<tr className="line">
				{this.props.lineNumber && <td className="line-number" data-line={this.props.lineNumber}></td>}
				<td className="line-content">
					{this.props.tokens.map((token, i) => {
						if (!token["URL"]) {
							return <span className={token.Class || ""} key={i}>{token.Label}</span>;
						}

						let cls = `${token.Class || ""} ref`;
						if (token.IsDef) {
							cls += " def";
						}
						switch (token.URL[0]) {
						case this.props.selectedDef:
							cls += " highlight-primary";
							break;
						case this.props.highlightedDef:
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
					{this.props.tokens.length === 0 && <span>&nbsp;</span>}
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
