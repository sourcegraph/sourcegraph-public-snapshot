import React from "react";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

// TODO support for tokens with more than one URL
class CodeLineView extends React.Component {
	componentWillMount() {
		this._updateOwnURLs(this.props.tokens);
	}

	componentWillReceiveProps(nextProps) {
		if (nextProps.tokens !== this.props.tokens) {
			this._updateOwnURLs(nextProps.tokens);
		}
	}

	shouldComponentUpdate(nextProps, nextState) {
		if (this.props.selectedDef !== null && nextState.ownURLs[this.props.selectedDef]) {
			return true;
		}
		if (nextProps.selectedDef !== null && nextState.ownURLs[nextProps.selectedDef]) {
			return true;
		}
		if (this.props.highlightedDef !== null && nextState.ownURLs[this.props.highlightedDef]) {
			return true;
		}
		if (nextProps.highlightedDef !== null && nextState.ownURLs[nextProps.highlightedDef]) {
			return true;
		}
		return false;
	}

	_updateOwnURLs(tokens) {
		let ownURLs = {};
		tokens.forEach((token) => {
			if (token.URL !== undefined) {
				ownURLs[token.URL[0]] = true;
			}
		});
		this.setState({ownURLs: ownURLs});
	}

	render() {
		return (
			<tr className="line">
				<td className="line-number" data-line={this.props.lineNumber}></td>
				<td className="line-content">
					{this.props.tokens.map((token, i) => {
						if (token.URL === undefined) {
							return <span className={token.Class} key={i}>{token.Label}</span>;
						}

						let cls = `${token.Class} ref`;
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

export default CodeLineView;
