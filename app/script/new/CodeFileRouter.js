import React from "react";
import URI from "urijs";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import CodeFileContainer from "./CodeFileContainer";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
export default class CodeFileRouter extends Component {
	componentDidMount() {
		this.dispatcherToken = Dispatcher.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
	}

	reconcileState(state, props) {
		state.uri = URI.parse(props.location);
		state.navigate = props.navigate || null;

		let pathParts = state.uri.path.substr(1).split("/.");
		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = repoParts[1];

		let keys = [];
		let vars = URI.parseQuery(state.uri.query);
		pathParts.slice(1).forEach((part) => {
			let p = part.indexOf("/");
			let key = part.substr(0, p);
			keys.push(key);
			vars[key] = part.substr(p + 1);
		});

		state.tree = vars["tree"] || null;
		state.startLine = vars["startline"] ? parseInt(vars["startline"], 10) : null;
		state.endLine = vars["endline"] ? parseInt(vars["endline"], 10) : null;
		state.selectedDef = vars["seldef"] || null;

		state.def = vars["def"] || null;
		state.unitType = state.def && keys[0];
		state.unit = state.def && vars[keys[0]];
		state.example = vars["examples"] ? parseInt(vars["examples"], 10) : null;
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

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.SelectLine:
			this._navigate(null, {
				startline: action.line,
				endline: action.line,
			});
			break;

		case CodeActions.SelectRange:
			this._navigate(null, {
				startline: Math.min(this.state.startLine || action.line, action.line),
				endline: Math.max(this.state.endLine || action.line, action.line),
			});
			break;

		case DefActions.SelectDef:
			// null becomes undefined
			this._navigate(null, {seldef: action.url || undefined}); // eslint-disable-line no-undefined
			break;

		case DefActions.GoToDef:
			console.warn("GoToDef: not yet implemented");
			break;
		}
	}

	render() {
		if (this.state.def) {
			return (
				<CodeFileContainer
					repo={this.state.repo}
					rev={this.state.rev}
					unitType={this.state.unitType}
					unit={this.state.unit}
					def={this.state.def}
					example={this.state.example} />
			);
		}

		return (
			<CodeFileContainer
				repo={this.state.repo}
				rev={this.state.rev}
				tree={this.state.tree}
				startLine={this.state.startLine}
				endLine={this.state.endLine}
				selectedDef={this.state.selectedDef}
				def={null} />
		);
	}
}

CodeFileRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};
