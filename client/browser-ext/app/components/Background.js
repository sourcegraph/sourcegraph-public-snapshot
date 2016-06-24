import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import {useAccessToken} from "../actions/xhr";
import * as Actions from "../actions";
import styles from "./App.css";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

import * as utils from "../utils";

let createdReposCache = {};

@connect(
	(state) => ({
		accessToken: state.accessToken,
		def: state.def,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class Background extends React.Component {
	static propTypes = {
		accessToken: React.PropTypes.string,
		def: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._refresh = this._refresh.bind(this);
		this._cleanupAndRefresh = this._cleanupAndRefresh.bind(this);
		this._popstateUpdate = this._popstateUpdate.bind(this);
		this._clickRef = this._clickRef.bind(this);
		this._directURLToDef = this._directURLToDef.bind(this);
		this._updateIntervalID = null;
	}

	componentDidMount() {
		if (this.props.accessToken) useAccessToken(this.props.accessToken);

		// Capture user's access token if on sourcegraph.com.
		if (utils.isSourcegraphURL()) {
			const regexp = /accessToken\\":\\"([-A-Za-z0-9_.]+)\\"/;
			const matchResult = document.head.innerHTML.match(regexp);
			if (matchResult) this.props.actions.setAccessToken(matchResult[1]);
		}

		if (this._updateIntervalID === null) {
			this._updateIntervalID = setInterval(this._refreshVCS.bind(this), 1000 * 60 * 5); // refresh every 5min
		}

		document.addEventListener("click", this._clickRef);
		document.addEventListener("pjax:success", this._cleanupAndRefresh);
		window.addEventListener("popstate", this._popstateUpdate);

		this._refresh();
	}

	componentWillUpdate(nextProps) {
		this._refresh();
	}

	componentWillUnmount() {
		document.removeEventListener("pjax:success", this._cleanupAndRefresh);
		document.removeEventListener("popstate", this._popstateUpdate);
		document.removeEventListener("click", this._clickRef);
		if (this._updateIntervalID !== null) {
			clearInterval(this._updateIntervalID);
			this._updateIntervalID = null;
		}
	}

	_clickRef(ev) {
		if (ev.target.dataset && typeof ev.target.dataset.sourcegraphRef !== "undefined") {
			let currLocation = utils.parseURLWithSourcegraphDef();
			let urlProps = utils.parseURLWithSourcegraphDef({pathname: ev.target.pathname, hash: ev.target.hash});
			this.props.actions.getDef(urlProps.repoURI, urlProps.rev, urlProps.defPath);

			const directURLToDef = this._directURLToDef(urlProps);
			if (directURLToDef) {
				EventLogger.logEvent("ClickedDef", {defPath: urlProps.defPath, repo: urlProps.repoURI, user: urlProps.user, direct: "true"});
				ev.target.href = `${directURLToDef.pathname}${directURLToDef.hash}`;
				this._renderDefInfo(this.props, urlProps);
			} else {
				EventLogger.logEvent("ClickedDef", {defPath: urlProps.defPath, repo: urlProps.repoURI, user: urlProps.user, direct: "false"});
				pjaxGoTo(ev.target.href, urlProps.repoURI === currLocation.repoURI);
			}
		}
	}

	_cleanupAndRefresh() {
		// Clean up any popovers on the page before refreshing (after pjax:success).
		// Otherwise, popovers may remain on the page because the anchored elem's mousout
		// event may not have fired (and the elem may no longer be on the page).
		const popovers = document.querySelectorAll(".sg-popover")
		for (let i = 0; i < popovers.length; ++i) {
			popovers[i].remove();
		}

		this._refresh();
	}

	_popstateUpdate() {
		// If the user navigates "back" in the browser, there will not necessarily
		// be a pjax:success event; it may be that the user is jumping back to
		// a previous definition (even in the same file) in which case re-rendering
		// the def info link is necessary.
		this._renderDefInfo(this.props, utils.parseURLWithSourcegraphDef());
	}

	_refresh() {
		// First, get the current browser state (which could have been updated by another tab).
		chrome.runtime.sendMessage(null, {type: "get"}, {}, (state) => {
			const accessToken = state.accessToken;
			if (accessToken) this.props.actions.setAccessToken(accessToken);

			if (utils.isSourcegraphURL()) return;

			let urlProps = utils.parseURLWithSourcegraphDef();

			// TODO: Branches that are not built on Sourcegraph will not get annotations, need to trigger
			if (urlProps.repoURI) {
				this.props.actions.refreshVCS(urlProps.repoURI);
			}
			if (urlProps.path) {
				// Strip hash (e.g. line location) from path.
				const hashLoc = urlProps.path.indexOf("#");
				if (hashLoc !== -1) urlProps.path = urlProps.path.substring(0, hashLoc);
			}

			if (urlProps.repoURI && urlProps.defPath && !urlProps.isDelta) {
				this.props.actions.getDef(urlProps.repoURI, urlProps.rev, urlProps.defPath);
			}

			if (urlProps.repoURI && !createdReposCache[urlProps.repoURI]) {
				createdReposCache[urlProps.repoURI] = true;
				this.props.actions.ensureRepoExists(urlProps.repoURI);
			}

			const directURLToDef = this._directURLToDef(urlProps);
			if (directURLToDef) {
				if (!window.location.href.includes(directURLToDef.hash)) {
					pjaxGoTo(`${directURLToDef.pathname}${directURLToDef.hash}`, true);
				}
			}

			this._renderDefInfo(this.props, urlProps);
		});
	}

	_refreshVCS() {
		let urlProps = utils.parseURLWithSourcegraphDef();
		if (urlProps.repoURI && utils.isGitHubURL()) {
			this.props.actions.refreshVCS(urlProps.repoURI);
		}
	}

	_directURLToDef({repoURI, rev, defPath}) {
		const defObj = this.props.def.content[keyFor(repoURI, rev, defPath)];
		if (defObj) {
			const pathname = `/${repoURI.replace("github.com/", "")}/${defObj.File === "." ? "tree" : "blob"}/${rev}/${defObj.File}`;
			const hash = `#sourcegraph&def=${defPath}&L${defObj.StartLine || 0}-${defObj.EndLine || 0}`;
			return {pathname, hash};
		}
		return null;
	}

	_renderDefInfo(props, state) {
		const def = props.def.content[keyFor(state.repoURI, state.rev, state.defPath)];

		const id = "sourcegraph-def-info";
		let e = document.getElementById(id);

		// Hide when no def is present.
		if (!def) {
			if (e) {
				e.remove();
			}
			return;
		}

		if (!e) {
			e = document.createElement("td");
			e.id = id;
			e.className = styles["def-info"];
			e.style.position = "absolute";
			e.style.right = "0";
			e.style.zIndex = "1000";
			e.style["-webkit-user-select"] = "none";
			e.style["user-select"] = "none";
		}
		let a = e.firstChild;
		if (!a) {
			a = document.createElement("a");
			e.appendChild(a);
		}

		a.href = `https://sourcegraph.com/${state.repoURI}@${state.rev}/-/info/${state.defPath}?utm_source=browser-ext&browser_type=chrome`;
		a.dataset.content = "Find Usages";
		a.target = "tab";
		a.title = `Sourcegraph: View cross-references to ${def.Name}`;

		// Anchor to def's start line.
		let anchor = document.getElementById(`L${def.StartLine}`);
		if (!anchor) {
			return;
		}
		anchor = anchor.parentNode;
		anchor.style.position = "relative";
		anchor.appendChild(e);
	}

	render() {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}

// pjaxGoTo uses GitHub's existing PJAX to navigate to a URL. It
// is faster than a hard page reload.
function pjaxGoTo(url, sameRepo) {
	if (!sameRepo) {
		window.location.href = url;
		return;
	}

	const e = document.createElement("a");
	e.href = url;
	if (sameRepo) e.dataset.pjax = "#js-repo-pjax-container";
	if (sameRepo) e.classList.add("js-navigation-open");
	document.body.appendChild(e);
	e.click();
	setTimeout(() => document.body.removeChild(e), 1000);
}
