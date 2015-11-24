import React from "react";
import ReactDOM from "react-dom";
import URI from "urijs";

import Component from "../Component";
import LocationAdaptor from "../LocationAdaptor";
import SearchResultsRouter from "./SearchResultsRouter";

export default class SearchBar extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchViewIsActive: false,
		};
		this._submitSearch = this._submitSearch.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.uri = URI.parse(props.location);
		state.navigate = props.navigate || null;

		let pathParts = state.uri.path.substr(1).split("/.");
		if (pathParts.length >= 2 && pathParts[1] === "search") {
			state.searchViewIsActive = true;
		} else {
			state.searchViewIsActive = false;
		}

		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = repoParts[1] || "master";

		let vars = URI.parseQuery(state.uri.query);
		state.query = vars["q"] || null;
	}

	onStateTransition(prevState, nextState) {
		if (!prevState.searchViewIsActive && nextState.searchViewIsActive) {
			ReactDOM.render((
				<LocationAdaptor component={SearchResultsRouter} />
			), document.getElementById("main"));
		} else if (prevState.searchViewIsActive && !nextState.searchViewIsActive) {
			window.location.href = URI.build(nextState.uri);
		}

		if (prevState.query !== nextState.query && this.refs.queryInput) {
			this.refs.queryInput.value = nextState.query;
		}
	}

	_navigate(path, query) {
		let uri = Object.assign({}, this.state.uri);
		if (path) {
			uri.path = path;
		}
		if (query) {
			uri.query = URI.buildQuery(Object.assign(URI.parseQuery(uri.query), query));
		}
		this.state.navigate(URI.build(uri));
	}

	_searchPath() {
		return `/${this.state.repo}@${this.state.rev}/.search`;
	}

	_submitSearch(e) {
		if (e) e.preventDefault();
		let query = this.refs.queryInput.value;
		this._navigate(this._searchPath(), {
			q: query,
			type: "tokens",
		});
	}

	render() {
		return (
			<form className="navbar-form" onSubmit={this._submitSearch.bind(this)}>
				<div className="form-group">
					<div className="input-group">
						<input className="form-control search-input-next"
							ref="queryInput"
							name="q"
							placeholder="Search"
							defaultValue={this.state.query} />
						<span className="input-group-addon"></span>
					</div>
				</div>
			</form>
		);
	}
}
