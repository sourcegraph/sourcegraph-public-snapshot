import React from "react";
import {render} from "react-dom";
import {bindActionCreators} from "redux";
import {connect, Provider} from "react-redux";

import addAnnotations from "./annotations";

import {useAccessToken} from "../../app/actions/xhr";
import * as Actions from "../../app/actions";
import Root from "../../app/containers/Root";
import {SearchIcon} from "../../app/components/Icons";
import {keyFor, getExpiredSrclibDataVersion, getExpiredDefs, getExpiredAnnotations} from "../../app/reducers/helpers";
import createStore from "../../app/store/configureStore";

@connect(
	(state) => ({
		accessToken: state.accessToken,
		repo: state.repo,
		rev: state.rev,
		path: state.path,
		srclibDataVersion: state.srclibDataVersion,
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
		srclibDataVersion: React.PropTypes.object.isRequired,
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
			window.addEventListener("focus", this.refreshState);
		}

		this.refreshState();
		document.addEventListener("pjax:success", this.pjaxUpdate);

		getExpiredSrclibDataVersion(this.props.srclibDataVersion).forEach(({repo, rev, path}) => this.props.actions.expireSrclibDataVersion(repo, rev, path));
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
	}

	componentWillUnmount() {
		document.removeEventListener("keydown", this.keyboardEvents);
		document.removeEventListener("pjax:success", this.pjaxUpdate);
		window.removeEventListener("focus", this.refreshState);
	}

	parseURL() {
		// TODO: this method has problems handling branch revisions with "/" character.
		const urlsplit = window.location.href.split("/");
		let user = urlsplit[3];
		let repo = urlsplit[4]
		let rev = "master"
		if (urlsplit[6] !== null && (urlsplit[5] === "tree" || urlsplit[5] === "blob")) { // what about "commit"
			rev = urlsplit[6];
		}
		let path = urlsplit.slice(7).join("/");
		return {user, repo, rev, path};
	}

	viewingGoBlob() {
		if (!this.props.path) return false;

		const pathParts = this.props.path.split("/");
		const lang = pathParts[pathParts.length - 1].split(".")[1] || null;
		return window.location.href.split("/")[5] === "blob" && document.querySelector(".file") && lang && lang.toLowerCase() === "go";
	}

	// refreshState is called whenever this component is mounted or
	// pjax completes successfully; it updates the store with the
	// current repo/rev/path. It will render navbar search button
	// (if none exists) and annotations for the current code file (if any).
	refreshState() {
		this.addSearchButton();

		let {user, repo, rev, path} = this.parseURL();
		if (repo) repo = `github.com/${user}/${repo}`;
		if (path) {
			// Strip hash (e.g. line location) from path.
			const hashLoc = path.indexOf("#");
			if (hashLoc !== -1) path = path.substring(0, hashLoc);
		}

		this.props.actions.setRepoRev(repo, rev);
		this.props.actions.setPath(path);
		if (repo && rev && this.viewingGoBlob()) {
			this.props.actions.getAnnotations(repo, rev, path);
		}

		const srclibDataVersion = this.props.srclibDataVersion.content[keyFor(repo, rev, path)];
		if (srclibDataVersion && srclibDataVersion.CommitID) {
			const annotations = this.props.annotations.content[keyFor(repo, srclibDataVersion.CommitID, path)];
			if (annotations) this.annotate(annotations);
		}
	}

	// pjaxUpdate is a wrapper around refresh state which is called whenever
	// pjax completes successfully, etc. It will also remove the app frame.
	pjaxUpdate() {
		this.removeAppFrame();
		this.refreshState();
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
			chrome.storage.local.get("state", (obj) => {
				const {state} = obj;
				const initialState = JSON.parse(state || "{}");

				const createStore = require("../../app/store/configureStore");

				const frameDiv = document.createElement("div");
				frameDiv.id="sourcegraph-frame";
				render(<Root store={createStore(initialState)} />, frameDiv);

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
		if (this.frameDiv) this.frameDiv.style.display = "none";
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
			document.querySelector(".repository-content").style.display = "none"
			this.frameDiv.style.display = "block";
			this.setState({appFrameIsVisible: true}, focusInput);
		}
	};

	annotate(json) {
		if (!this.viewingGoBlob()) return;

		let fileElem = document.querySelector(".file .blob-wrapper");
		if (fileElem) {
			if (document.querySelector(".vis-private") && !this.props.accessToken) {
				console.error("Sourcegraph chrome extension will not work on private code until you login on Sourcegraph.com");
			} else {
				addAnnotations(json);
			}
		}
	}

	render() {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}

const bootstrapApp = function() {
	chrome.storage.local.get("state", obj => {
		const {state} = obj;
		const initialState = JSON.parse(state || "{}");

		const app = document.createElement("div");
		app.id = "sourcegraph-app-bootstrap";
		app.style.display = "none";
		render(<Provider store={createStore(initialState)}><InjectApp /></Provider>, app);

		document.body.appendChild(app);
	});
}

window.addEventListener("load", bootstrapApp);
