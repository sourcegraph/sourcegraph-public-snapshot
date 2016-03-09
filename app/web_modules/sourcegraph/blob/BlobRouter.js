import React from "react";
import URL from "url";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import BlobContainer from "sourcegraph/blob/BlobContainer";
import DefStore from "sourcegraph/def/DefStore";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
import {GoTo} from "sourcegraph/util/hotLink";

function lineCol(line, col) {
	if (typeof col === "undefined") {
		return line.toString();
	}
	return `${line}:${col}`;
}

function lineRange(startLineCol, endLineCol) {
	if (typeof endLineCol === "undefined" || startLineCol === endLineCol) {
		return startLineCol;
	}
	return `${startLineCol}-${endLineCol}`;
}

function parseLineRange(range) {
	let lineMatch = range.match(/^(\d+)(?:-(\d+))?$/);
	if (lineMatch) {
		return {
			startLine: parseInt(lineMatch[1], 10),
			endLine: parseInt(lineMatch[2] || lineMatch[1], 10),
		};
	}
	let lineColMatch = range.match(/^(\d+):(\d+)-(\d+):(\d+)$/);
	if (lineColMatch) {
		return {
			startLine: parseInt(lineColMatch[1], 10),
			startCol: parseInt(lineColMatch[2], 10),
			endLine: parseInt(lineColMatch[3], 10),
			endCol: parseInt(lineColMatch[4], 10),
		};
	}
}

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
class BlobRouter extends Component {
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

		if (!state.url.pathname.match(/\/\.(tree|def)/)) {
			// Not a file view. Could be a new URL navigated to by the search bar.
			return;
		}

		state.tree = null;
		state.def = null;
		state.startLine = null;
		state.endLine = null;
		if (pathParts[1].startsWith("tree/")) {
			state.tree = pathParts[1].slice("tree/".length);

			if (state.url.hash) {
				if (state.url.hash.startsWith("#L")) {
					Object.assign(state, parseLineRange(state.url.hash.slice(2)));
				}
			}
		} else {
			state.def = state.url.pathname;
		}
	}

	_navigate(pathname, o) {
		let hash;
		if (o && o.startLine) {
			hash = `L${lineRange(lineCol(o.startLine, o.startCol), o.endLine && lineCol(o.endLine, o.endCol))}`;
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
		case BlobActions.SelectLine:
			this._navigate(this._filePath(), {startLine: action.line});
			break;

		case BlobActions.SelectLineRange:
			this._navigate(this._filePath(), {
				startLine: Math.min(this.state.startLine || action.line, action.line),
				endLine: Math.max(this.state.endLine || this.state.startLine || action.line, action.line),
			});
			break;

		case BlobActions.SelectCharRange:
			this._navigate(this._filePath(), {
				startLine: action.startLine,
				startCol: action.startCol,
				endLine: action.endLine,
				endCol: action.endCol,
			});
			break;

		case DefActions.SelectDef:
			{
				let def = DefStore.defs.get(action.url);
				if (def) {
					if (!def.Error) {
						this._navigate(action.url);
					}
				} else {
					this.setState({loadingDef: action.url});
				}
				break;
			}

		case DefActions.DefFetched:
			if (this.state.loadingDef === action.url) {
				this._navigate(action.url);
				this.setState({loadingDef: null});
			}
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
		return (
			<BlobContainer
				repo={this.state.repo}
				rev={this.state.rev}
				tree={this.state.tree}
				startLine={this.state.startLine || null}
				startCol={this.state.startCol || null}
				endLine={this.state.endLine || null}
				endCol={this.state.endCol || null}
				activeDef={this.state.def} />
		);
	}
}

BlobRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};

export default BlobRouter;
