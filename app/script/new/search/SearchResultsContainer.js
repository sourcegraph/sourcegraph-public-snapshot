import React from "react";

import Container from "../Container";
import Dispatcher from "../Dispatcher";
import SearchResultsStore from "./SearchResultsStore";
import * as SearchActions from "./SearchActions";
import "./SearchBackend";

export default class SearchResultsContainer extends Container {
	stores() {
		return [SearchResultsStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.results = SearchResultsStore.results;
		state.resultsGeneration = SearchResultsStore.results.generation;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.query !== prevState.query) {
			for (let type of resultTypes) {
				Dispatcher.asyncDispatch(
					new SearchActions.WantResults(nextState.repo, nextState.rev, type.type, 1, type.perPage, nextState.query)
				);
			}
		}
	}

	render() {
		let results = this.state.results.get(this.state.repo, this.state.rev, this.state.type, 1);
		return <p>{results && results.Total}</p>;
	}
}

class ResultType {
	constructor(type, name, perPage, component) {
		this.type = type;
		this.name = name;
		this.perPage = perPage;
		this.component = component;
	}
}

const resultTypes = [
	new ResultType("tokens", "Definitions", 50, <p>Tokens</p>),
	new ResultType("text", "Text", 10, <p>Text</p>),
];

SearchResultsContainer.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	type: React.PropTypes.string,
	query: React.PropTypes.string,
};
