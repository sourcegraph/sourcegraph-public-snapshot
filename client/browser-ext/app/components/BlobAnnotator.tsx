import {allActions} from "../actions";
import {EventLogger} from "../analytics/EventLogger";
import {AnnotationsState, ResolvedRevState} from "../reducers";
import {keyFor} from "../reducers/helpers";
import * as utils from "../utils";
import {addAnnotations} from "../utils/annotations";
import * as github from "../utils/github";
import {SourcegraphIcon} from "./Icons";
import * as React from "react";
import {connect} from "react-redux";
import {bindActionCreators} from "redux";

const isCloning = new Set<String>();

interface Props {
	path: string;
	repoURI: string;
	blobElement: HTMLElement;
	selfElement: HTMLElement;
}

interface ReduxProps {
	actions: typeof allActions;
	resolvedRev: ResolvedRevState;
	annotations: AnnotationsState;
}

class Base extends React.Component<Props & ReduxProps, {}> {
	refreshInterval: NodeJS.Timer;

	// language is determined by the path extension
	language: string;

	isDelta?: boolean;
	isCommit?: boolean;
	isPullRequest?: boolean;
	isSplitDiff?: boolean;

	// rev is defined for blob view
	rev?: string;

	// base/head properties are defined for diff views (commit + pull request)
	baseCommitID?: string;
	headCommitID?: string;
	baseBranch?: string;
	headBranch?: string;
	baseRepoURI?: string;
	headRepoURI?: string;

	constructor(props: Props & ReduxProps) {
		super(props);
		this.language = utils.getPathExtension(props.path);

		this._clickRefresh = this._clickRefresh.bind(this);

		const {isDelta, isPullRequest, isCommit, rev} = utils.parseURL(window.location);
		this.isDelta = isDelta;
		this.isPullRequest = isPullRequest;
		this.isCommit = isCommit;
		this.rev = rev;

		if (this.isDelta) {
			this.isSplitDiff = github.isSplitDiff();
			const deltaRevs = github.getDeltaRevs();
			if (!deltaRevs) {
				// TODO(john): error handling
				return;
			}

			this.baseCommitID = deltaRevs.base;
			this.headCommitID = deltaRevs.head;

			const deltaInfo = github.getDeltaInfo();
			if (!deltaInfo) {
				// TODO(john): error handling
				return;
			}

			this.baseRepoURI = deltaInfo.baseURI;
			this.headRepoURI = deltaInfo.headURI;
		}

		if (this.baseRepoURI !== this.headRepoURI && this.headRepoURI) {
			// Ensure the head repo of a cross-repo PR is created.
			props.actions.ensureRepoExists(this.headRepoURI);
		}

		this.fetchAnnotations();
		this._addAnnotations();
	}

	componentDidMount(): void {
		github.registerExpandDiffClickHandler(this._clickRefresh);
	}

	componentDidUpdate(): void {
		// Reapply annotations after reducer state changes.
		this._addAnnotations();
	}

	_clickRefresh(): void {
		// Diff expansion is not synchronous, so we must wait for
		// elements to get added to the DOM before calling into the
		// annotations code. 500ms is arbitrary but seems to work well.
		setTimeout(() => this._addAnnotations(), 500);
	}

	fetchAnnotations(): void {
		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				this.props.actions.getAnnotations(this.baseRepoURI, this.baseCommitID, this.props.path);
			}
			if (this.headCommitID && this.headRepoURI) {
				this.props.actions.getAnnotations(this.headRepoURI, this.headCommitID, this.props.path);
			}
		} else if (this.rev) {
			this.props.actions.getAnnotations(this.props.repoURI, this.rev, this.props.path);
		}
	}

	_addAnnotations(): void {
		const apply = (repoURI: string, rev: string, isBase: boolean, loggerProps: Object) => {
			const fext = utils.getPathExtension(this.props.path);

			if (utils.supportedExtensions.indexOf(fext) < 0) {
				return; // Don't annotate unsupported languages
			}

			const json = this.props.annotations.content[keyFor(repoURI, rev, this.props.path)];
			if (json) {
				addAnnotations(this.props.path, {repoURI, rev, isDelta: this.isDelta || false, isBase, relRev: json.relRev}, this.props.blobElement, json.resp.IncludedAnnotations.Annotations, json.resp.IncludedAnnotations.LineStartBytes, this.isSplitDiff || false, loggerProps);
			}
		};

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				apply(this.baseRepoURI, this.baseCommitID, true, this.eventLoggerProps());
			}
			if (this.headCommitID && this.headRepoURI) {
				apply(this.headRepoURI, this.headCommitID, false, this.eventLoggerProps());
			}
		} else {
			const resolvedRev = this.props.resolvedRev.content[keyFor(this.props.repoURI, this.rev)];
			if (resolvedRev && resolvedRev.json && resolvedRev.json.CommitID) {
				apply(this.props.repoURI, resolvedRev.json.CommitID, false, this.eventLoggerProps());
			}
		}
	}

	eventLoggerProps(): Object {
		return {
			repoURI: this.props.repoURI,
			path: this.props.path,
			isPullRequest: this.isPullRequest,
			isCommit: this.isCommit,
			language: this.language,
			isPrivateRepo: github.isPrivateRepo(),
		};
	}

	getBlobUrl(): string {
		return `https://sourcegraph.com/${this.props.repoURI}${this.rev ? `@${this.rev}` : ""}/-/blob/${this.props.path}`;
	}

	render(): JSX.Element | null {
		if (typeof this.props.resolvedRev.content[keyFor(this.props.repoURI)] === "undefined") {
			return null;
		}

		if (github.isPrivateRepo() && this.props.resolvedRev.content[keyFor(this.props.repoURI)].authRequired) {
			// Not signed in or not auth'd for private repos
			this.props.selfElement.removeAttribute("disabled");
			this.props.selfElement.setAttribute("aria-label", `Authorize Sourcegraph for ${this.props.repoURI.split("github.com/")[1]}`);

			return (<span>
				<a href={`https://sourcegraph.com/authext?rtg=${encodeURIComponent(window.location.href)}`}
					style={{textDecoration: "none", color: "inherit"}}>
					<SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px", WebkitFilter: "grayscale(100%)"}} />
					Sourcegraph
				</a>
			</span>);

		} else if (this.props.resolvedRev.content[keyFor(this.props.repoURI)].cloneInProgress) {
			// Cloning the repo
			this.props.selfElement.setAttribute("disabled", "true");
			this.props.selfElement.setAttribute("aria-label", `Sourcegraph is analyzing ${this.props.repoURI.split("github.com/")[1]}`);

			if (!isCloning.has(this.props.repoURI)) {
				isCloning.add(this.props.repoURI);
				this.refreshInterval = setInterval(this.fetchAnnotations, 5000);
			}

			return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Loading...</span>;

		} else if (!utils.supportedExtensions.includes(utils.getPathExtension(this.props.path))) {
			this.props.selfElement.setAttribute("disabled", "true");
			if (!utils.upcomingExtensions.includes(utils.getPathExtension(this.props.path))) {
				this.props.selfElement.setAttribute("aria-label", "File not supported");
			} else {
				this.props.selfElement.setAttribute("aria-label", "Language support coming soon!");
			}

			return <span style={{pointerEvents: "none"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />Sourcegraph</span>;

		} else {
			this.props.selfElement.removeAttribute("disabled");
			this.props.selfElement.setAttribute("aria-label", "View on Sourcegraph");

			if (isCloning.has(this.props.repoURI)) {
				isCloning.delete(this.props.repoURI);
				if (this.refreshInterval) {
					clearInterval(this.refreshInterval);
				}
			}

			return (<span>
				<a href={this.getBlobUrl()} style={{textDecoration: "none", color: "inherit"}}><SourcegraphIcon style={{marginTop: "-1px", paddingRight: "4px", fontSize: "18px"}} />
					Sourcegraph
				</a>
			</span>);
		}
	}
}

export const BlobAnnotator = connect((state) => ({
	resolvedRev: state.resolvedRev,
	annotations: state.annotations,
}), (dispatch) => ({actions: bindActionCreators(allActions, dispatch)}))(Base) as React.ComponentClass<Props>;
