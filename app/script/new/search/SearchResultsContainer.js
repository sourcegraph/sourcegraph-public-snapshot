import React from "react";

import Container from "../Container";
import SearchResultsStore from "./SearchResultsStore";

export default class SearchResultsContainer extends Container {
	stores() {
		return [SearchResultsStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return <p>{this.state.query}</p>;
	}
}

SearchResultsContainer.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	query: React.PropTypes.string,
};
