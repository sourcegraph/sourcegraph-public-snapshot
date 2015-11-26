import React from "react";

import Container from "../Container";
import Dispatcher from "../Dispatcher";
import SearchResultsStore from "./SearchResultsStore";
import * as SearchActions from "./SearchActions";
import TokenSearchResults from "./TokenSearchResults";
import TextSearchResults from "./TextSearchResults";
import "./SearchBackend";

class ResultType {
	constructor(label, name, perPage, component) {
		this.label = label;
		this.name = name;
		this.perPage = perPage;
		this.component = component;
	}
}

const resultTypes = [
	new ResultType("tokens", "Definitions", 50, TokenSearchResults),
	new ResultType("text", "Text", 10, TextSearchResults),
];

export default class SearchResultsContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			currentType: resultTypes[0],
		};
	}

	stores() {
		return [SearchResultsStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.results = SearchResultsStore.results;
		state.resultsGeneration = SearchResultsStore.results.generation;
		state.currentType = resultTypes.find((type) => type.label === props.type);
	}

	onStateTransition(prevState, nextState) {
		if (nextState.query !== prevState.query) {
			for (let type of resultTypes) {
				Dispatcher.asyncDispatch(
					new SearchActions.WantResults(nextState.repo, nextState.rev, type.label, 1, type.perPage, nextState.query)
				);
			}
		}
	}

	render() {
		let currentResult = this.state.results.get(this.state.repo, this.state.rev, this.state.query, this.state.currentType.label, 1);

		return (
			<div className="search-results row">
				<div className="col-md-10 col-md-offset-1">
					<ul className="nav nav-pills">
						{resultTypes.map((type) => {
							let results = this.state.results.get(this.state.repo, this.state.rev, this.state.query, type.label, 1);
							return (
								<li key={type.label} className={type.label === this.state.currentType.label ? "active" : null}>
									<a onClick={() => {
										Dispatcher.dispatch(new SearchActions.SelectResultType(type.label));
									}}>
										<i className="fa fa-asterisk"></i> {type.name} <span className="badge">{results ? results.Total : <i className="fa fa-circle-o-notch fa-spin"></i>}</span>
									</a>
								</li>
							);
						})}
					</ul>
					{currentResult &&
						<this.state.currentType.component
							repo={this.state.repo}
							rev={this.state.rev}
							query={this.state.query}
							resultData={currentResult} />
					}
				</div>
			</div>
		);
	}
}

SearchResultsContainer.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	type: React.PropTypes.string,
	query: React.PropTypes.string,
};
