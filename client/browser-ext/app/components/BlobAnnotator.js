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

const isCloning = new Map();

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

		this.getBlobUrl = this.getBlobUrl.bind(this);
		this._lineRefresh = this._lineRefresh.bind(this);
		this._hashRefresh = this._hashRefresh.bind(this);
		this._clickRefresh = this._clickRefresh.bind(this);
		this.onClickAuthPriv = this.onClickAuthPriv.bind(this);
		this.onClickFileView = this.onClickFileView.bind(this);
		this.refreshWhenVCSCloned = this.refreshWhenVCSCloned.bind(this);

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

			this._lineRefresh();
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

		window.addEventListener("hashchange", this._hashRefresh);
	}

	componentWillUnmount() {
		let diffExpanders = document.getElementsByClassName("diff-expander");
		if (diffExpanders) {
			for (let idx = 0; idx < diffExpanders.length; idx++) {
				diffExpanders[idx].removeEventListener("click", this._clickRefresh);
			}
		}

		window.removeEventListener("hashchange", this._hashRefresh);
	}

	_clickRefresh() {
		// Diff expansion is not synchronous, so we must wait for
		// elements to get added to the DOM before calling into the
		// annotations code. 500ms is arbitrary but seems to work well.
		setTimeout(() => {
			this._lineRefresh();
			this._hashRefresh();
			this._addAnnotations(this.state);

			// Attach to the new diff expander;
			// TODO: Only attach to the new diff-expander
			let diffExpanders = document.getElementsByClassName("diff-expander");
			if (diffExpanders) {
				for (let idx = 0; idx < diffExpanders.length; idx++) {
					diffExpanders[idx].addEventListener("click", this._clickRefresh);
				}
			}
		}, 500);
	}

	_lineRefresh() {
		// Get first line number of the file's first head hunk. In a split diff, there
		// is an extra element before the element containing the head data-line-number,
		// which is why we use a different nth-child value.
		if (!this.state.isDelta) return;

		let headLineNumberEl = this.props.blobElement.querySelector(`[data-line-number]:nth-child(${this.state.isSplitDiff ? 3 : 2})`);
		if (headLineNumberEl) {
			this.state.headLineNumber = headLineNumberEl.dataset.lineNumber;
		}
	}

	_hashRefresh() {
		// Do not modify the URL if not auth'd or if not supported
		if ((this.isPrivateRepo() &&
			(typeof this.state.resolvedRev.content[keyFor(this.state.repoURI)] !== 'undefined') &&
			(typeof this.state.resolvedRev.content[keyFor(this.state.repoURI)].authRequired === 'undefined')) ||
			(utils.supportedExtensions.includes(utils.getPathExtension(this.state.path)))) {

			const selfElementA = this.state.selfElement.getElementsByTagName("A");
			if (selfElementA) {
				selfElementA.href = this.getBlobUrl();
			}
		}
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
				addAnnotations(state.path, {repoURI, rev, branch, isDelta: state.isDelta, isBase, relRev: json.relRev}, state.blobElement, json.resp.IncludedAnnotations.Annotations, json.resp.IncludedAnnotations.LineStartBytes, state.isSplitDiff);
			}
		}

		if (state.isDelta) {
			if (state.baseCommitID) apply(state.baseRepoURI, state.baseCommitID, state.base, true);
			if (state.headCommitID) apply(state.headRepoURI, state.headCommitID, state.head, false);
		} else {
			const resolvedRev = state.resolvedRev.content[keyFor(state.repoURI, state.rev)];
			if (resolvedRev && resolvedRev.json && resolvedRev.json.CommitID) apply(state.repoURI, resolvedRev.json.CommitID, state.rev, false);
		}
	}

	isPrivateRepo() {
		const el = document.getElementsByClassName("label label-private v-align-middle");
		return el.length > 0
	}

	onClickAuthPriv(ev) {
		ev.preventDefault();
		EventLogger.logEventForCategory("Auth", "Redirect", "ChromeExtensionSgButtonClicked", {repo: this.state.repoURI, path: window.location.href, is_private_repo: this.isPrivateRepo()});
		location.href = `https://sourcegraph.com/authext?rtg=${encodeURIComponent(window.location.href)}`;
	}

	onClickFileView(ev) {
		ev.preventDefault();
		EventLogger.logEventForCategory("File", "Click", "ChromeExtensionSgButtonClicked", {repo: this.state.repoURI, path: window.location.href, is_private_repo: this.isPrivateRepo()});
		const targetURL = this.getBlobUrl();
		if (ev.ctrlKey || (navigator.platform.toLowerCase().indexOf('mac') >= 0 && ev.metaKey) || ev.button === 1) {
			// Remove :focus from target to remove the hover
			// tooltip when opening target link in a new window.
			ev.target.blur();
			window.open(targetURL, "_blank");
		} else {
			location.href = targetURL;
		}
	}

	getBlobUrl() {
		let rev;
		let repo;
		let selectedLineNumber = "";

		switch(utils.getGitHubRoute()) {
			case "blob":
				rev = this.state.rev;
				repo = this.state.repoURI;
				selectedLineNumber = `${window.location.hash.match(/^(#L[1-9][0-9]*$)/g) || ""}`;
				break;
			case "pull":
			case "commit":
				const selectedLine = this.state.blobElement.getElementsByClassName("js-linkable-line-number selected-line")[0];
				const jumpToSide = selectedLine ? `${selectedLine.id.match(/([LR])[\d]+$/g) || "R"}` : "R";

				// If no line is selected, jump to first line in the condensed blob
				selectedLineNumber = selectedLine && selectedLine.dataset ? `#L${selectedLine.dataset.lineNumber}` : "";
				if (!selectedLineNumber && this.state.headLineNumber) {
					selectedLineNumber = `#L${this.state.headLineNumber}`;
				}

				if (jumpToSide.charAt(0) === "L") {
					rev = this.state.baseCommitID;
					repo = this.state.baseRepoURI;
				} else {
					rev = this.state.headCommitID;
					repo = this.state.headRepoURI;
				}

				break;
			default:
				console.debug("Could not generate blob URL");
				return window.location;
		}

		return `https://sourcegraph.com/${repo}@${rev}/-/blob/${this.state.path}${selectedLineNumber}`;
	}

	refreshWhenVCSCloned() {
		this.state.actions.getAnnotations(this.state.repoURI, this.state.rev, this.state.path);
	}


	render() {
		if (typeof this.state.resolvedRev.content[keyFor(this.state.repoURI)] !== "undefined") {
			if (this.isPrivateRepo() && this.state.resolvedRev.content[keyFor(this.state.repoURI)].authRequired === true) {
				// Not signed in or not auth'd for private repos
				this.state.selfElement.removeAttribute("disabled");
				this.state.selfElement.setAttribute("aria-label", `Authorize Sourcegraph for ${this.state.repoURI.split("github.com/")[1]}`);
				this.state.selfElement.onclick = this.onClickAuthPriv;

				return <span><a href={`https://sourcegraph.com/authext?rtg=${encodeURIComponent(window.location.href)}`} onclick={this.onClickAuthPriv} style={{textDecoration: "none", color: "inherit"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px", WebkitFilter: "grayscale(100%)"}} />Sourcegraph</a></span>;
			} else if (this.state.resolvedRev.content[keyFor(this.state.repoURI)].cloneInProgress === true) {
				// Cloning the repo
				this.state.selfElement.setAttribute("disabled", true);
				this.state.selfElement.setAttribute("aria-label", `Sourcegraph is analyzing ${this.state.repoURI.split("github.com/")[1]}`);

				if (isCloning.has(this.state.repoURI) === false) {
					isCloning.set(this.state.repoURI, true);
					this.state.refreshInterval = setInterval(this.refreshWhenVCSCloned, 5000);
				}

				return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Loading...</span>;
			} else {
				if (!utils.supportedExtensions.includes(utils.getPathExtension(this.state.path))) {
					this.state.selfElement.setAttribute("disabled", true);
					if (!utils.upcomingExtensions.includes(utils.getPathExtension(this.state.path))) {
						this.state.selfElement.setAttribute("aria-label", "File not supported");
					} else {
						this.state.selfElement.setAttribute("aria-label", "Language support coming soon!");
					}

					return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Sourcegraph</span>;
				} else {
					this.state.selfElement.removeAttribute("disabled");
					this.state.selfElement.setAttribute("aria-label", "View on Sourcegraph");
					this.state.selfElement.onclick = this.onClickFileView;

					if (isCloning.has(this.state.repoURI) === true) {
						isCloning.delete(this.state.repoURI);
						if (this.state.refreshInterval)
							clearInterval(this.state.refreshInterval);
					}

					return <span><a id="SourcegraphFileViewAnchor" href={this.getBlobUrl()} onclick={this.onClickFileView} style={{textDecoration: "none", color: "inherit"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Sourcegraph</a></span>;
				}
			}
		} else {
			// Default case when we don't have any annotation data
			return <span><a id="SourcegraphFileViewAnchor" href={this.getBlobUrl()} onclick={this.onClickFileView} style={{textDecoration: "none", color: "inherit"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Sourcegraph</a></span>;
		}
	}
}
