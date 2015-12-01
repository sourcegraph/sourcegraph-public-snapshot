import React from "react";
import URI from "urijs";

import Component from "../Component";
import Dispatcher from "../Dispatcher";
import CodeFileContainer from "./CodeFileContainer";
import DefStore from "../def/DefStore";
import * as CodeActions from "./CodeActions";
import * as DefActions from "../def/DefActions";
import {GoTo} from "../util/hotLink";

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
class CodeFileRouter extends Component {
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
		state.rev = repoParts[1] || "master";

		// We split the URI path based on `/.` because that usually denotes an
		// operation, but in the case of the tree operation consider this path:
		//
		//  "/sourcegraph@master/.tree/.gitignore"
		//  "/sourcegraph@master/.tree/subdirectory/.gitignore"
		//
		// In the above, .gitignore is the file name not the operation. So we handle
		// this case specially here.
		if (pathParts.length >= 2 && (pathParts[1] === "tree" || pathParts[1].indexOf("tree/") === 0)) {
			// Parse the filepath following "/.tree/".
			let treePath = state.uri.path.substring(state.uri.path.indexOf("/.tree/") + "/.tree/".length);

			// Reform the pathParts array with the corrected URI path split.
			pathParts = [pathParts[0], `tree/${treePath}`];
		}

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

		state.def = vars["def"] ? `/${state.repo}@${state.rev}/.${keys[0]}/${vars[keys[0]]}/.def/${vars["def"]}` : null;
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
			this._navigate(this._filePath(), {
				startline: action.line,
				endline: action.line,
			});
			break;

		case CodeActions.SelectRange:
			this._navigate(this._filePath(), {
				startline: Math.min(this.state.startLine || action.line, action.line),
				endline: Math.max(this.state.endLine || action.line, action.line),
			});
			break;

		case DefActions.SelectDef:
			// null becomes undefined
			this._navigate(this._filePath(), {seldef: action.url || undefined}); // eslint-disable-line no-undefined
			break;

		case GoTo:
			this.state.navigate(URI(action.url).absoluteTo(this.state.uri).href());
			break;
		}
	}

	_filePath() {
		let tree = this.state.tree || DefStore.defs.get(this.state.def).File.Path;
		return `/${this.state.repo}@${this.state.rev}/.tree/${tree}`;
	}

	render() {
		if (this.state.def) {
			return (
				<CodeFileContainer
					repo={this.state.repo}
					rev={this.state.rev}
					def={this.state.def} />
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

export default CodeFileRouter;
