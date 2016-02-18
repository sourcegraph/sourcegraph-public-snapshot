import React from "react";
import URL from "url";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import CodeFileContainer from "sourcegraph/code/CodeFileContainer";
import DefStore from "sourcegraph/def/DefStore";
import * as CodeActions from "sourcegraph/code/CodeActions";
import * as DefActions from "sourcegraph/def/DefActions";
import {GoTo} from "sourcegraph/util/hotLink";

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
		state.url = URL.parse(props.location, true);
		state.navigate = props.navigate || null;

		let pathParts = state.url.pathname.substr(1).split("/.");
		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = decodeURIComponent(repoParts[1] || "");

		// We split the path based on `/.` because that usually denotes an
		// operation, but in the case of the tree operation consider this path:
		//
		//  "/sourcegraph@master/.tree/.gitignore"
		//  "/sourcegraph@master/.tree/subdirectory/.gitignore"
		//
		// In the above, .gitignore is the file name not the operation. So we handle
		// this case specially here.
		if (pathParts.length >= 2 && (pathParts[1] === "tree" || pathParts[1].indexOf("tree/") === 0)) {
			// Parse the filepath following "/.tree/".
			let treePath = state.url.pathname.substring(state.url.pathname.indexOf("/.tree/") + "/.tree/".length);

			// Reform the pathParts array with the corrected path split.
			pathParts = [pathParts[0], `tree/${treePath}`];
		}

		state.tree = null;
		state.def = null;
		state.startLine = null;
		state.endLine = null;
		state.selectedDef = null;
		if (pathParts[1].startsWith("tree/")) {
			state.tree = pathParts[1].slice("tree/".length);

			if (state.url.hash) {
				let lineMatch = state.url.hash.match(/^#L(\d+)(?:-(\d+))?$/);
				state.startLine = lineMatch ? parseInt(lineMatch[1], 10) : null;
				if (lineMatch && lineMatch[2]) {
					state.endLine = parseInt(lineMatch[2], 10);
				} else {
					state.endLine = state.startLine;
				}

				let defMatch = state.url.hash.match(/^#def-(.+)$/);
				state.selectedDef = defMatch ? defMatch[1] : null;
			}
		} else {
			state.def = state.url.pathname;
		}
	}

	_navigate(pathname, startLine, endLine, def) {
		let hash;
		if (startLine && endLine && startLine !== endLine) {
			hash = `L${startLine}-${endLine}`;
		} else if (startLine) {
			hash = `L${startLine}`;
		} else if (def) {
			hash = `def-${def}`;
		}
		let url = {
			protocol: this.state.url.protocol,
			auth: this.state.url.auth,
			host: this.state.url.host,
			pathname: pathname || this.state.url.pathname,
			hash: hash,
		};
		this.state.navigate(URL.format(url));
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case CodeActions.SelectLine:
			this._navigate(this._filePath(), action.line);
			break;

		case CodeActions.SelectRange:
			this._navigate(
				this._filePath(),
				Math.min(this.state.startLine || action.line, action.line),
				Math.max(this.state.endLine || action.line, action.line)
			);
			break;

		case DefActions.SelectDef:
			// null becomes undefined
			this._navigate(this._filePath(), null, null, action.url); // eslint-disable-line no-undefined
			break;

		case GoTo:
			this.state.navigate(URL.resolve(this.state.url, action.url));
			break;
		}
	}

	_filePath() {
		let tree = this.state.tree || DefStore.defs.get(this.state.def).File.Path;
		let revPart = this.state.rev ? `@${this.state.rev}` : "";
		return `/${this.state.repo}${revPart}/.tree/${tree}`;
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
