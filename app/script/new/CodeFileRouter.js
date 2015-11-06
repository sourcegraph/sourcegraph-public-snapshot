import React from "react";
import URI from "urijs";

import Dispatcher from "./Dispatcher";
import CodeFileContainer from "./CodeFileContainer";
import * as DefActions from "./DefActions";

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
export default class CodeFileRouter extends React.Component {
	constructor(props, context) {
		super(props, context);
	}

	componentDidMount() {
		this.dispatcherToken = Dispatcher.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
	}

	_navigate(path, query) {
		let uri = URI.parse(this.props.location);
		if (path) {
			uri.path = path;
		}
		if (query) {
			uri.query = URI.buildQuery(Object.assign(URI.parseQuery(uri.query), query));
		}
		this.props.navigate(URI.build(uri));
	}

	__onDispatch(action) {
		switch (action.constructor) {
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
		let uri = URI.parse(this.props.location);
		let pathParts = uri.path.substr(1).split("/.");

		let keys = [];
		let vars = URI.parseQuery(uri.query);

		let repoParts = pathParts[0].split("@");
		vars["repo"] = repoParts[0];
		vars["rev"] = repoParts[1];

		pathParts.slice(1).forEach((part) => {
			let p = part.indexOf("/");
			let key = part.substr(0, p);
			keys.push(key);
			vars[key] = part.substr(p + 1);
		});

		if (vars["def"]) {
			return (
				<CodeFileContainer
					repo={vars["repo"]}
					rev={vars["rev"]}
					unitType={keys[0]}
					unit={vars[keys[0]]}
					def={vars["def"]}
					example={vars["examples"] ? parseInt(vars["examples"], 10) : null} />
			);
		}

		return (
			<CodeFileContainer
				repo={vars["repo"]}
				rev={vars["rev"]}
				tree={vars["tree"]}
				startLine={vars["startline"] ? parseInt(vars["startline"], 10) : null}
				endLine={vars["endline"] ? parseInt(vars["endline"], 10) : null}
				selectedDef={vars["seldef"] || null}
				def={null} />
		);
	}
}

CodeFileRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};
