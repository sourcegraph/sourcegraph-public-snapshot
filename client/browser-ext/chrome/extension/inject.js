import React from "react";
import {render} from "react-dom";
import {bindActionCreators} from "redux";
import {connect, Provider} from "react-redux";

import addAnnotations from "./annotations";

import {useAccessToken} from "../../app/actions/xhr";
import * as Actions from "../../app/actions";
import Root from "../../app/containers/Root";
import styles from "../../app/components/App.css";
import {SearchIcon} from "../../app/components/Icons";
import {keyFor, getExpiredSrclibDataVersion, getExpiredDef, getExpiredDefs, getExpiredAnnotations} from "../../app/reducers/helpers";
import createStore from "../../app/store/configureStore";

@connect(
	(state) => ({
		accessToken: state.accessToken,
		repo: state.repo,
		rev: state.rev,
		path: state.path,
		defPath: state.defPath,
		srclibDataVersion: state.srclibDataVersion,
		def: state.def,
		annotations: state.annotations,
		defs: state.defs,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
class InjectApp extends React.Component {
	static propTypes = {
		accessToken: React.PropTypes.string,
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
		path: React.PropTypes.string,
		defPath: React.PropTypes.string,
		srclibDataVersion: React.PropTypes.object.isRequired,
		def: React.PropTypes.object.isRequired,
		annotations: React.PropTypes.object.isRequired,
		defs: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired
	};

	constructor(props) {
		super(props);
		this.state = {
			appFrameIsVisible: false,
		};
		this.refreshState = this.refreshState.bind(this);
		this.keyboardEvents = this.keyboardEvents.bind(this);
		this.removeAppFrame = this.removeAppFrame.bind(this);
		this.toggleAppFrame = this.toggleAppFrame.bind(this);
		this.pjaxUpdate = this.pjaxUpdate.bind(this);
		this.focusUpdate = this.focusUpdate.bind(this);
		this._clickRef = this._clickRef.bind(this);
	}

	componentDidMount() {
		if (this.props.accessToken) useAccessToken(this.props.accessToken);

		// Capture the access token if on sourcegraph.com.
		if (window.location.href.match(/https:\/\/(www.)?sourcegraph.com/)) {
			const regexp = /accessToken\\":\\"([-A-Za-z0-9_.]+)\\"/;
			const matchResult = document.head.innerHTML.match(regexp);
			if (matchResult) this.props.actions.setAccessToken(matchResult[1]);
		}

		if (window.location.href.match(/https:\/\/(www.)?github.com/)) {
			document.addEventListener("keydown", this.keyboardEvents);

			// The window focus listener will refresh state to reflect the
			// current repository being viewed.
			window.addEventListener("focus", this.focusUpdate);
		}

		this.refreshState();
		document.addEventListener("pjax:success", this.pjaxUpdate);
		document.addEventListener("click", this._clickRef);

		getExpiredSrclibDataVersion(this.props.srclibDataVersion).forEach(({repo, rev, path}) => this.props.actions.expireSrclibDataVersion(repo, rev, path));
		getExpiredDef(this.props.def).forEach(({repo, rev, defPath}) => this.props.actions.expireDef(repo, rev, defPath));
		getExpiredDefs(this.props.defs).forEach(({repo, rev, path, query}) => this.props.actions.expireDefs(repo, rev, path, query));
		getExpiredAnnotations(this.props.annotations).forEach(({repo, rev, path}) => this.props.actions.expireAnnotations(repo, rev, path));
	}

	componentWillReceiveProps(nextProps) {
		// Annotation data is fetched asynchronously; annotate the page if the new props
		// contains annotation data for the current blob.
		const srclibDataVersion = nextProps.srclibDataVersion.content[keyFor(nextProps.repo, nextProps.rev, nextProps.path)];
		if (srclibDataVersion && srclibDataVersion.CommitID) {
			const annotations = nextProps.annotations.content[keyFor(nextProps.repo, srclibDataVersion.CommitID, nextProps.path)];
			if (annotations) this.annotate(annotations);
		}

		// Show/hide def info.
		if (nextProps.defPath && (nextProps.repo !== this.props.repo || nextProps.rev !== this.props.rev || nextProps.defPath !== this.props.defPath || nextProps.def !== this.props.def)) {
			this._renderDefInfo(nextProps);
		}
	}

	componentWillUnmount() {
		document.removeEventListener("keydown", this.keyboardEvents);
		document.removeEventListener("pjax:success", this.pjaxUpdate);
		window.removeEventListener("focus", this.focusUpdate);
		document.removeEventListener("click", this._clickRef);
	}

	_clickRef(ev) {
		if (typeof ev.target.dataset.sourcegraphRef !== "undefined" || (ev.target.parentNode && typeof ev.target.parentNode.dataset.sourcegraphRef !== "undefined")) {
			let urlProps = this.parseURL({pathname: ev.target.pathname, hash: ev.target.hash});
			urlProps.repo = `github.com/${urlProps.user}/${urlProps.repo}`;

			this.props.actions.getDef(urlProps.repo, urlProps.rev, urlProps.defPath);

			const props = {...urlProps, def: this.props.def};
			const info = this._directURLToDef(props);
			if (info) {
				// Fast path. Uses PJAX if possible (automatically).
				const {pathname, hash} = info;
				ev.target.href = `${pathname}${hash}`;
				this._renderDefInfo(props);
			} else {
				pjaxGoTo(ev.target.href, urlProps.repo === this.props.repo);
			}
		}
	}

	parseURL(loc = window.location) {
		// TODO: this method has problems handling branch revisions with "/" character.
		const urlsplit = loc.pathname.slice(1).split("/");
		let user = urlsplit[0];
		let repo = urlsplit[1]
		// We scrape the current branch and set rev to it so we stay on the same branch when doing jump-to-def
		let currBranch = document.getElementsByClassName('select-menu-button js-menu-target css-truncate')[0].title
		let rev = currBranch
		if (urlsplit[3] !== null && (urlsplit[2] === "tree" || urlsplit[2] === "blob")) { // what about "commit"
			rev = urlsplit[3];
		}
		let path = urlsplit.slice(4).join("/");

		const info = {user, repo, rev, path};

		// Check for URL hashes like "#sourcegraph&def=...".
		if (loc.hash.startsWith("#sourcegraph&")) {
			const parts = loc.hash.slice(1).split("&").slice(1); // omit "sourcegraph" sentinel
			parts.forEach((p) => {
				const kv = p.split("=", 2);
				if (kv.length != 2) return;
				let k = kv[0];
				const v = kv[1];
				if (k === "def") k = "defPath"; // disambiguate with def obj
				if (!info[k]) info[k] = v; // don't clobber
			});
		}
		return info;
	}

	supportsAnnotatingFile(path) {
		if (!path) return false;

		const pathParts = path.split("/");
		let lang = pathParts[pathParts.length - 1].split(".")[1] || null;
		lang = lang ? lang.toLowerCase() : null;
		const supportedLang = lang === "go" || lang === "java";
		return window.location.href.split("/")[5] === "blob" && document.querySelector(".file") && supportedLang;
	}

	// refreshState is called whenever this component is mounted or
	// pjax completes successfully; it updates the store with the
	// current repo/rev/path. It will render navbar search button
	// (if none exists) and annotations for the current code file (if any).
	refreshState() {
		this.addSearchButton();

		let {user, repo, rev, path, defPath} = this.parseURL();
		// This scrapes the latest commit ID and updates rev to the latest commit so we are never injecting
		// outdated annotations.  If there is a new commit, srclib-data-version will return a 404, but the
		// refresh endpoint will update the version and the annotations will be up to date once the new build succeeds
		let latestRev = document.getElementsByClassName("js-permalink-shortcut")[0].href.split("/")[6]
		// TODO: Branches that are not built on Sourcegraph will not get annotations, need to trigger
		let currBranch = document.getElementsByClassName('select-menu-button js-menu-target css-truncate')[0].title
		if (rev !== latestRev) rev = latestRev;
		if (currBranch !== "master") rev = currBranch;
		const repoName = repo;
		if (repo) {
			repo = `github.com/${user}/${repo}`;
			this.props.actions.refreshVCS(repo);
		}
		if (path) {
			// Strip hash (e.g. line location) from path.
			const hashLoc = path.indexOf("#");
			if (hashLoc !== -1) path = path.substring(0, hashLoc);
		}

		this.props.actions.setRepoRev(repo, rev);
		this.props.actions.setDefPath(defPath);
		this.props.actions.setPath(path);

		if (repo && defPath) {
			this.props.actions.getDef(repo, rev, defPath);
		}

		if (repo && rev && this.supportsAnnotatingFile(path)) {
			this.props.actions.getAnnotations(repo, rev, path);
		}

		this._renderDefInfo(this.props);

		const srclibDataVersion = this.props.srclibDataVersion.content[keyFor(repo, rev, path)];
		if (srclibDataVersion && srclibDataVersion.CommitID) {
			const annotations = this.props.annotations.content[keyFor(repo, srclibDataVersion.CommitID, path)];
			if (annotations) this.annotate(annotations);
		}
	}

	// _checkNavigateToDef checks for a URL fragment of the form "#sourcegraph&def=..."
	// and redirects to the def's definition in code on GitHub.com.
	_checkNavigateToDef({repo, rev, defPath, def}) {
		const info = this._directURLToDef({repo, rev, defPath, def});
		if (info) {
			const {pathname, hash} = info;
			if (!(window.location.pathname === pathname && window.location.hash === hash)) {
				pjaxGoTo(`${pathname}${hash}`, repo === this.props.repo);
			}
		}
	}

	_directURLToDef({repo, rev, defPath, def}) {
		const defObj = def ? def.content[keyFor(repo, rev, defPath)] : null;
		if (defObj) {
			const pathname = `/${repo.replace("github.com/", "")}/blob/${rev}/${defObj.File}`;
			const hash = `#sourcegraph&def=${defPath}&L${defObj.StartLine || 0}-${defObj.EndLine || 0}`;
			return {pathname, hash};
		}
		return null;
	}

	// pjaxUpdate is a wrapper around refreshState which is called whenever
	// pjax completes successfully, etc. It will also remove the app frame.
	pjaxUpdate() {
		this.removeAppFrame();
		this.refreshState();
	}

	// focusUpdate is a wrapper around refreshState which is called whenever
	// the window tab becomes focused on GitHub.com; it will first read
	// local storage for any data (e.g. Sourcegraph access token) set via other
	// tabs.
	focusUpdate() {
		chrome.runtime.sendMessage(null, {type: "get"}, {}, (state) => {
			const accessToken = state.accessToken;
			if (accessToken) this.props.actions.setAccessToken(accessToken); // without this, access token may be overwritten to null
			this.refreshState();
		});
	}

	// addSearchButton injects a button into the GitHub pagehead actions bar
	// (next to "watch" and "star" and "fork" actions). It is idempotent
	// but the injected component is separated from the react component
	// hierarchy.
	addSearchButton() {
		let pagehead = document.querySelector("ul.pagehead-actions");
		if (pagehead && !pagehead.querySelector("#sg-search-button-container")) {
			let button = document.createElement("li");
			button.id = "sg-search-button-container";

			render(
				// this button inherits styles from GitHub
				<button className="btn btn-sm minibutton tooltipped tooltipped-s" aria-label="Keyboard shortcut: shift-T" onClick={this.toggleAppFrame}>
					<SearchIcon /><span style={{paddingLeft: "5px"}}>Search code</span>
				</button>, button
			);
			pagehead.insertBefore(button, pagehead.firstChild);
		}
	}

	// appFrame creates a div frame embedding the chrome extension (react) app.
	// It can be injected into the DOM when desired. It is idempotent, i.e.
	// returns the (already mounted) DOM element if one has already been created.
	// It returns the div asynchronously, since the application bootstrap requires
	// (asynchronously) connecting to chrome local storage.
	appFrame(cb) {
		if (!this.frameDiv) {
			chrome.runtime.sendMessage(null, {type: "get"}, {}, (state) => {
				const createStore = require("../../app/store/configureStore");

				const frameDiv = document.createElement("div");
				frameDiv.id = "sourcegraph-frame";
				render(<Root store={createStore(state)} />, frameDiv);

				this.frameDiv = frameDiv;
				cb(frameDiv);
			});
		} else {
			cb(this.frameDiv);
		}
	}

	keyboardEvents(e) {
		if (e.which === 84 && e.shiftKey && (e.target.tagName.toLowerCase()) !== "input" && (e.target.tagName.toLowerCase()) !== "textarea" && !this.state.appFrameIsVisible) {
			this.toggleAppFrame();
		} else if (e.keyCode === 27 && this.state.appFrameIsVisible) {
			this.toggleAppFrame();
		}
	}

	removeAppFrame = () => {
		const el = document.querySelector(".repository-content");
		if (el) el.style.display = "block";
		const frame = document.getElementById("sourcegraph-frame");
		if (frame) frame.style.display = "none";
		this.setState({appFrameIsVisible: false});
	}

	// toggleAppFrame is the handler for the pagehead "search code" button;
	// it will directly manipulate the DOM to hide all GitHub repository
	// content and mount an iframe embedding the chrome extension (react) app.
	toggleAppFrame = () => {
		const focusInput = () => {
			const el = document.querySelector(".sg-input");
			if (el) setTimeout(() => el.focus()); // Auto focus input, with slight delay so T doesn't appear
		}

		if (!document.getElementById('sourcegraph-frame')) {
			// Lazy initial application bootstrap; add app frame to DOM.
			this.appFrame((frameDiv) => {
				document.querySelector(".repository-content").style.display = "none";
				document.querySelector(".container.new-discussion-timeline").appendChild(frameDiv);
				frameDiv.style.display = "block";
				this.setState({appFrameIsVisible: true}, focusInput);
			});
		} else if (this.state.appFrameIsVisible) {
			// Toggle visibility off.
			this.removeAppFrame();
		} else {
			// Toggle visiblity on.
			document.querySelector(".repository-content").style.display = "none";
			const frame = document.getElementById("sourcegraph-frame");
			if (frame) frame.style.display = "block";
			this.setState({appFrameIsVisible: true}, focusInput);
		}
	};

	annotate(json) {
		let fileElem = document.querySelector(".file .blob-wrapper");
		if (fileElem) {
			if (document.querySelector(".vis-private") && !this.props.accessToken) {
				console.error("To use the Sourcegraph Chrome extension on private code, sign in at https://sourcegraph.com and add your repositories.");
			} else {
				addAnnotations(json);
			}
		}
	}

	_renderDefInfo(props) {
		const def = props.def.content[keyFor(props.repo, props.rev, props.defPath)];

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

		a.href = `https://sourcegraph.com/${props.repo}/-/info/${props.defPath}`;
		a.dataset.content = "Find Usages";
		a.target = "tab";
		a.title = `Sourcegraph: View cross-references to ${def.Name}`;

		// Anchor to def's start line.
		let anchor = document.getElementById(`L${def.StartLine}`);
		if (!anchor) {
			console.error("no line number element to anchor def info to");
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

const bootstrapApp = function() {
	chrome.runtime.sendMessage(null, {type: "get"}, {}, (state) => {
		const app = document.createElement("div");
		app.id = "sourcegraph-app-bootstrap";
		app.style.display = "none";
		render(<Provider store={createStore(state)}><InjectApp /></Provider>, app);

		document.body.appendChild(app);
	});
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

window.addEventListener("load", bootstrapApp);
