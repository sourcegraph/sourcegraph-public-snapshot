import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";
import addAnnotations from "../utils/annotations";
import Component from "./Component";
import {SourcegraphIcon} from "./Icons";
import * as Actions from "../actions";
import * as utils from "../utils";
import {keyFor} from "../reducers/helpers";
import EventLogger from "../analytics/EventLogger";

let buildsCache = {};

@connect(
	(state) => ({
		resolvedRev: state.resolvedRev,
		annotations: state.annotations,
		accessToken: state.accessToken,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class BlobAnnotator extends Component {
	static propTypes = {
		path: React.PropTypes.string.isRequired,
		resolvedRev: React.PropTypes.object.isRequired,
		annotations: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
		blobElement: React.PropTypes.object.isRequired,
		infoElement: React.PropTypes.object.isRequired,
		selfElement: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this._clickRefresh = this._clickRefresh.bind(this);
		this.onClickAuthPriv = this.onClickAuthPriv.bind(this);
		this.onClickFileView = this.onClickFileView.bind(this);

		this.state = utils.parseURL();
		this.state.path = props.path;

		if (this.state.isDelta) {
			this.state.isSplitDiff = this._isSplitDiff();

			let baseCommitID, headCommitID;
			let el = document.getElementsByClassName("js-socket-channel js-updatable-content js-pull-refresh-on-pjax");
			if (el && el.length > 0) {
				// Blob view
				for (let i = 0; i < el.length; ++i) {
					const url = el[i].dataset ? el[i].dataset.url : null;
					if (!url) continue;
					const urlSplit = url.split("?");
					if (urlSplit.length !== 2) continue;
					const query = urlSplit[1];
					const querySplit = query.split("&");
					for (let kv of querySplit) {
						const kvSplit = kv.split("=");
						const k = kvSplit[0];
						const v = kvSplit[1];
						if (k === "base_commit_oid") baseCommitID = v;
						if (k === "end_commit_oid") headCommitID = v;
					}
				}
			} else if (props.infoElement.tagName === 'A') {
				// For snippets in conversation view of pull request
				const baseInput = document.querySelector('input[name="comparison_base_oid"]');
				if (baseInput) {
					baseCommitID = baseInput.value;
				}
				if (typeof props.infoElement.href !== 'undefined') {
					headCommitID = props.infoElement.href.split("/files/")[1].split("#diff")[0];
				} else {
					const headInput = document.querySelector('input[name="comparison_end_oid"]');
					if (headInput) {
						headCommitID = headInput.value;
					}
				}
			} else if (this.state.isPullRequest) {
				// Files changed view in pull requests
				const baseInput = document.querySelector('input[name="comparison_base_oid"]');
				if (baseInput) {
					baseCommitID = baseInput.value;
				}
				const headInput = document.querySelector('input[name="comparison_end_oid"]');
				if (headInput) {
					headCommitID = headInput.value;
				}
			} else if (this.state.isCommit) {
				// Files changed view in commits
				let shaContainer = document.querySelectorAll(".sha-block");
				if (shaContainer && shaContainer.length === 2) {
					let baseShaEl = shaContainer[0].querySelector("a");
					if (baseShaEl) baseCommitID = baseShaEl.href.split("/").slice(-1)[0];
					let headShaEl = shaContainer[1].querySelector("span.sha");
					if (headShaEl) headCommitID = headShaEl.innerText;
				}
			}
			if (!baseCommitID) {
				console.error("unable to parse base commit id");
			}
			if (!headCommitID) {
				console.error("unable to parse head commit id");
			}

			this.state.baseCommitID = baseCommitID;
			this.state.headCommitID = headCommitID;

			if (this.state.isPullRequest) {
				const branches = document.querySelectorAll(".commit-ref,.current-branch");
				this.state.base = branches[0].innerText;
				this.state.head = branches[1].innerText;

				if (this.state.base.includes(":")) {
					const baseSplit = this.state.base.split(":");
					this.state.base = baseSplit[1];
					this.state.baseRepoURI = `github.com/${baseSplit[0]}/${this.state.repo}`;
				} else {
					this.state.baseRepoURI = this.state.repoURI;
				}
				if (this.state.head.includes(":")) {
					const headSplit = this.state.head.split(":");
					this.state.head = headSplit[1];
					this.state.headRepoURI = `github.com/${headSplit[0]}/${this.state.repo}`;
				} else {
					this.state.headRepoURI = this.state.repoURI;
				}
			} else if (this.state.isCommit) {
				let branchEl = document.querySelector("li.branch");
				if (branchEl) branchEl = branchEl.querySelector("a")
				if (branchEl) {
					this.state.base = branchEl.innerText;
					this.state.head = branchEl.innerText;
				}
				this.state.baseRepoURI = this.state.repoURI;
				this.state.headRepoURI = this.state.repoURI;
			}

			// Get first line number of the file's first head hunk. In a split diff, there
			// is an extra element before the element containing the head data-line-number,
			// which is why we use a different nth-child value.
			let headLineNumberEl = props.blobElement.querySelector(`[data-line-number]:nth-child(${this.state.isSplitDiff ? 3 : 2})`);
			if (headLineNumberEl) {
				this.state.headLineNumber = headLineNumberEl.dataset.lineNumber;
			}
		}

		if (this.state.baseRepoURI !== this.state.headRepoURI) {
			// Ensure the head repo of a cross-repo PR is created.
			props.actions.ensureRepoExists(this.state.headRepoURI);
		}
	}

	componentDidMount() {
		// Handle context expansion, in which case we should
		// re-annotate the blob (which is smart enough to only annoate
		// lines which haven't already been annotated).
		let diffExpanders = document.getElementsByClassName("diff-expander");
		if (diffExpanders) {
			for (let idx = 0; idx < diffExpanders.length; idx++) {
				diffExpanders[idx].addEventListener("click", this._clickRefresh);
			}
		}
	}

	componentWillUnmount() {
		let diffExpanders = document.getElementsByClassName("diff-expander");
		if (diffExpanders) {
			for (let idx = 0; idx < diffExpanders.length; idx++) {
				diffExpanders[idx].removeEventListener("click", this._clickRefresh);
			}
		}
	}

	_clickRefresh() {
		// Diff expansion is not synchronous, so we must wait for
		// elements to get added to the DOM before calling into the
		// annotations code. 500ms is arbitrary but seems to work well.
		setTimeout(() => this._addAnnotations(this.state), 500);
	}

	_isSplitDiff() {
		if (this.state.isPullRequest) {
			const headerBar = document.getElementsByClassName("float-right pr-review-tools");
			if (!headerBar || headerBar.length !== 1) return false;

			const diffToggles = headerBar[0].getElementsByClassName("BtnGroup");
			if (!diffToggles || diffToggles.length !== 1) return false;

			const disabledToggle = diffToggles[0].getElementsByTagName("A")[0];
			return disabledToggle && !disabledToggle.href.includes("diff=split");
		} else {
			const headerBar = document.getElementsByClassName("details-collapse table-of-contents js-details-container");
			if (!headerBar || headerBar.length !== 1) return false;

			const diffToggles = headerBar[0].getElementsByClassName("BtnGroup float-right");
			if (!diffToggles || diffToggles.length !== 1) return false;

			const selectedToggle = diffToggles[0].querySelector(".selected");
			return selectedToggle && selectedToggle.href.includes("diff=split");
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		if (!state.isAnnotated) {
			if (state.isDelta) {
				if (state.baseCommitID) {
					state.actions.getAnnotations(state.baseRepoURI, state.baseCommitID, state.path);
				}
				if (state.headCommitID) {
					state.actions.getAnnotations(state.headRepoURI, state.headCommitID, state.path);
				}
				state.isAnnotated = true;
			} else {
				state.actions.getAnnotations(state.repoURI, state.rev, state.path);
				state.isAnnotated = true;
			}
		}
	}

	onStateTransition(prevState, nextState) {
		this._addAnnotations(nextState);
	}

	_addAnnotations(state) {
		function apply(repoURI, rev, branch, isBase) {
			const fext = utils.getPathExtension(state.path);

			if (utils.supportedExtensions.indexOf(fext) < 0) {
				return; // Don't annotate unsupported languages
			}

			const json = state.annotations.content[keyFor(repoURI, rev, state.path)];
			if (json) {
				addAnnotations(state.path, {repoURI, rev, branch, isDelta: state.isDelta, isBase}, state.blobElement, json.IncludedAnnotations.Annotations, json.IncludedAnnotations.LineStartBytes, state.isSplitDiff);
			}
		}

		if (state.isDelta) {
			if (state.baseCommitID) apply(state.baseRepoURI, state.baseCommitID, state.base, true);
			if (state.headCommitID) apply(state.headRepoURI, state.headCommitID, state.head, false);
		} else {
			const resolvedRev = state.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (resolvedRev && resolvedRev.CommitID) apply(state.repoURI, resolvedRev.CommitID, state.rev, false);
		}
	}

	isPrivateRepo() {
		const el = document.getElementsByClassName("label label-private v-align-middle");
		return el.length > 0
	}

	onClickAuthPriv(ev) {
		EventLogger.logEventForCategory("Auth", "Redirect", "ChromeExtensionSgButtonClicked", {repo: this.state.repoURI, path: window.location.href, is_private_repo: this.isPrivateRepo()});
		location.href = `https://sourcegraph.com/join?ob=github&rtg=${encodeURIComponent(window.location.href)}`;
	}

	onClickFileView(ev) {
		EventLogger.logEventForCategory("File", "Click", "ChromeExtensionSgButtonClicked", {repo: this.state.repoURI, path: window.location.href, is_private_repo: this.isPrivateRepo()});

		const repo = this.state.headRepoURI || this.state.repoURI;
		const rev = this.state.headCommitID || this.state.rev;
		const lineNumberFragment = this.state.headLineNumber ? `#L${this.state.headLineNumber}` : "";
		const targetURL = `https://sourcegraph.com/${repo}@${rev}/-/blob/${this.state.path}${lineNumberFragment}`;
		if (ev.ctrlKey || (navigator.platform.toLowerCase().indexOf('mac') >= 0 && ev.metaKey) || event.button !== 0) {
			// Remove :focus from target to remove the hover
			// tooltip when opening target link in a new window.
			ev.target.blur();
			window.open(targetURL, "_blank");
		} else {
			location.href = targetURL;
		}
	}

	render() {
		if (this.isPrivateRepo() &&
			(typeof this.state.resolvedRev.content[keyFor(this.state.repoURI)] !== 'undefined') &&
			(typeof this.state.resolvedRev.content[keyFor(this.state.repoURI)].authRequired !== 'undefined')) {

			// Not signed in or not auth'd for private repos
			this.state.selfElement.removeAttribute("disabled");
			this.state.selfElement.setAttribute("aria-label", `Authorize Sourcegraph for ${this.state.repoURI.split("github.com/")[1]}`);
			this.state.selfElement.onclick = this.onClickAuthPriv;
		} else {
			if (utils.supportedExtensions.includes(utils.getPathExtension(this.state.path))) {
				this.state.selfElement.setAttribute("aria-label", "View on Sourcegraph");
				this.state.selfElement.onclick = this.onClickFileView;

				return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Sourcegraph</span>;
			} else {
				// TODO: Only set style to disabled and log the click event for statistics on unsupported languages?
				this.state.selfElement.setAttribute("disabled", true);

				if (utils.upcomingExtensions.includes(utils.getPathExtension(this.state.path))) {
					this.state.selfElement.setAttribute("aria-label", "Language support coming soon!");
				} else {
					this.state.selfElement.setAttribute("aria-label", "File not supported");
				}
			}
		}

		return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px", WebkitFilter: "grayscale(100%)"}} />Sourcegraph</span>;
	}
}
