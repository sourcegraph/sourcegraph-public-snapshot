import React from "react";
import URL from "url";
import last from "lodash/array/last";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import BlobContainer from "sourcegraph/blob/BlobContainer";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefStore from "sourcegraph/def/DefStore";
import RepoStore from "sourcegraph/repo/RepoStore";
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
	constructor(props) {
		super(props);
		this.state = {
			// _isMounted is set by BlobRouter_test.js to test frontend behavior.
			_isMounted: Boolean(props._isMounted), // eslint-disable-line react/prop-types
		};
	}

	componentDidMount() {
		this.dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));

		// Normally this is a bad practice, but it's done out of neccessity with a good
		// reason here. We're rendering this component server-side, where the hash fragment
		// cannot be known. So on initial load, we render without the hash fragment taken
		// into consideration (both on server- and client-side), but after loading client-side,
		// we re-render with hash fragment used.
		this.setState({ // eslint-disable-line react/no-did-mount-set-state
			_isMounted: true,
		});
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
		this.setState({
			_isMounted: false,
		});
	}

	reconcileState(state, props) {
		state.url = URL.parse(props.location, true);
		state.navigate = props.navigate || null;

		let pathParts = state.url.pathname.substr(1).split("/.");
		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = decodeURIComponent(repoParts[1] || "");

		// If no branch is specified, aggressively resolve it to the default branch
		// so that any data fetched under this rev is more easily cacheable. (If we
		// didn't do this, then we'd have 2 cache entries for a lot of the same data:
		// one with the key including the default branch, and one with the key
		// having a rev of "").
		//
		// But don't actually block fetching on this. Usually we will have the repo
		// available synchronously because it was preloaded.
		if (state.rev === "") {
			let repoObj = RepoStore.repos.get(state.repo);
			if (repoObj) state.rev = repoObj.DefaultBranch;
		}

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
			// Not a file view.
			return;
		}

		state.path = null;
		state.def = null;
		state.startLine = null;
		state.endLine = null;
		state.viewRefs = false;
		if (pathParts[1].startsWith("tree/")) {
			state.path = pathParts[1].slice("tree/".length);

			if (state._isMounted) {
				if (state.url.hash) {
					if (state.url.hash.startsWith("#L")) {
						const range = parseLineRange(state.url.hash.slice(2));
						state.startLine = range.startLine || null;
						state.startCol = range.startCol || null;
						state.endLine = range.endLine || null;
						state.endCol = range.endCol || null;
					}
				}
			}
		} else {
			// TODO better way to do this routing.
			state.def = state.url.pathname.replace(/\/\.refs\/?$/, "");
			if (state.rev) {
				// If state.def is a rev-less def URL (referring to the repo's default branch),
				// make state.def contain the def URL with the rev, if the rev can be determined.
				// This ensures that the BlobContainer's activeDef is an exact string match to
				// the def's ref links, so that the def gets highlighted.
				const repoNoRevPrefix = `/${state.repo}/.`;
				if (state.def.startsWith(repoNoRevPrefix)) {
					state.def = `/${state.repo}@${state.rev}/.${state.def.slice(repoNoRevPrefix.length)}`;
				}
			}
			state.viewRefs = last(pathParts) === "refs";
			if (state.viewRefs) {
				state.path = state.url.query.Files ? state.url.query.Files : null;
			}
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
			this._navigate(this._filePath(action.repo, action.rev, action.path), {startLine: action.line});
			break;

		case BlobActions.SelectLineRange:
			this._navigate(this._filePath(action.repo, action.rev, action.path), {
				startLine: Math.min(this.state.startLine || action.line, action.line),
				endLine: Math.max(this.state.endLine || this.state.startLine || action.line, action.line),
			});
			break;

		case BlobActions.SelectCharRange:
			this._navigate(this._filePath(action.repo, action.rev, action.path), {
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

	_filePath(repo, rev, path) {
		if (!repo) repo = this.state.repo;
		if (!rev) rev = this.state.rev;
		if (!path) path = this.state.path || DefStore.defs.get(this.state.def).File;

		let revPart = rev ? `@${rev}` : "";
		return `/${this.state.repo}${revPart}/.tree/${path}`;
	}

	render() {
		// TODO more solid routing here.
		return (
			<div>
				{this.state.viewRefs &&
					<RefsContainer
						repo={this.state.repo}
						rev={this.state.rev}
						path={this.state.path}
						def={this.state.def} />
				}
				{!this.state.viewRefs &&
					<BlobContainer
						repo={this.state.repo}
						rev={this.state.rev}
						path={this.state.path}
						startLine={this.state.startLine || null}
						startCol={this.state.startCol || null}
						endLine={this.state.endLine || null}
						endCol={this.state.endCol || null}
						activeDef={this.state.def} />
				}
			</div>
		);
	}
}

BlobRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};

export default BlobRouter;
