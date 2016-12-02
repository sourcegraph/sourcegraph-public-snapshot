import * as backend from "../backend";
import * as utils from "../utils";
import { addAnnotations } from "../utils/annotations";
import * as github from "../utils/github";
import { SourcegraphIcon } from "./Icons";
import * as React from "react";

const className = "btn btn-sm tooltipped tooltipped-n";
const buttonStyle = { marginRight: "5px" };
const iconStyle = { marginTop: "-1px", paddingRight: "4px", fontSize: "18px" };

interface Props {
	path: string;
	repoURI: string;
	blobElement: HTMLElement;
}

interface State {
	resolvedRevs: { [key: string]: backend.ResolvedRevResp };
}

export class BlobAnnotator extends React.Component<Props, State> {
	revisionChecker: NodeJS.Timer;

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

	constructor(props: Props) {
		super(props);
		this.state = {
			resolvedRevs: {},
		};

		this.clickRefresh = this.clickRefresh.bind(this);
		this.updateResolvedRevs = this.updateResolvedRevs.bind(this);
		this.resolveRevs = this.resolveRevs.bind(this);
		this.hasUnresolvedRevs = this.hasUnresolvedRevs.bind(this);
		this.addAnnotations = this.addAnnotations.bind(this);
		this.eventLoggerProps = this.eventLoggerProps.bind(this);

		this.language = utils.getPathExtension(props.path);

		const {isDelta, isPullRequest, isCommit, rev} = utils.parseURL(window.location);
		this.isDelta = isDelta;
		this.isPullRequest = isPullRequest;
		this.isCommit = isCommit;
		this.rev = rev;

		if (this.isDelta) {
			this.isSplitDiff = github.isSplitDiff();
			const deltaRevs = github.getDeltaRevs();
			if (!deltaRevs) {
				console.error("cannot determine deltaRevs");
				return;
			}

			this.baseCommitID = deltaRevs.base;
			this.headCommitID = deltaRevs.head;

			const deltaInfo = github.getDeltaInfo();
			if (!deltaInfo) {
				console.error("cannot determine deltaInfo");
				return;
			}

			this.baseRepoURI = deltaInfo.baseURI;
			this.headRepoURI = deltaInfo.headURI;
		}

		if (this.baseRepoURI !== this.headRepoURI && this.headRepoURI) {
			// Ensure the head repo of a cross-repo PR is created.
			backend.ensureRepoExists(this.headRepoURI);
		}

		this.resolveRevs();
		this.addAnnotations();
	}

	componentDidMount(): void {
		github.registerExpandDiffClickHandler(this.clickRefresh);
		// Set a timer to re-check revision data every 10 seconds, for repos that haven't been
		// cloned and revs that haven't been sync'd to Sourcegraph.com.
		// Single-flighted requests / caching prevents spamming the API.
		this.revisionChecker = setInterval(() => this.resolveRevs(), 5000);
	}

	componentWillUnmount(): void {
		if (this.revisionChecker) {
			clearInterval(this.revisionChecker);
		}
	}

	componentDidUpdate(): void {
		// Reapply annotations after reducer state changes.
		this.addAnnotations();
	}

	clickRefresh(): void {
		// Diff expansion is not synchronous, so we must wait for
		// elements to get added to the DOM before calling into the
		// annotations code. 500ms is arbitrary but seems to work well.
		setTimeout(() => this.addAnnotations(), 500);
	}

	updateResolvedRevs(repo: string, rev?: string): void {
		const key = backend.cacheKey(repo, rev);
		if (this.state.resolvedRevs[key] && this.state.resolvedRevs[key].commitID) {
			return; // nothing to do
		}
		backend.resolveRev(repo, rev).then((resp) => {
			let repoStat;
			if (rev) {
				// Empty rev is checked to determine if the user has access to the repo.
				// Non-empty is checked to determine if Sourcegraph.com is sync'd.
				repoStat = { [repo]: resp };
			}
			this.setState({ resolvedRevs: Object.assign({}, this.state.resolvedRevs, { [key]: resp }, repoStat) });
		});
	}

	resolveRevs(): void {
		const repoStat = this.state.resolvedRevs[this.props.repoURI];
		if (repoStat && repoStat.notFound) {
			// User doesn't have permission to view repo; no need to fetch.
			return;
		}

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				this.updateResolvedRevs(this.baseRepoURI, this.baseCommitID);
			}
			if (this.headCommitID && this.headRepoURI) {
				this.updateResolvedRevs(this.headRepoURI, this.headCommitID);
			}
		} else if (this.rev) {
			this.updateResolvedRevs(this.props.repoURI, this.rev);
		} else {
			console.error("unable to fetch annotations; missing revision data");
		}
	}

	hasUnresolvedRevs(): boolean {
		const hasRev = (uri: string, rev?: string) => {
			const resolvedRev = this.state.resolvedRevs[backend.cacheKey(uri, rev)];
			return resolvedRev && resolvedRev.notFound;
		};

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI && !hasRev(this.baseRepoURI, this.baseCommitID)) {
				return true;
			}
			if (this.headCommitID && this.headRepoURI && !hasRev(this.headRepoURI, this.headCommitID)) {
				return true;
			}
		} else if (!hasRev(this.props.repoURI, this.rev)) {
			return true;
		}

		return false;
	}

	addAnnotations(): void {
		const apply = (repoURI: string, rev: string, isBase: boolean, loggerProps: Object) => {
			const ext = utils.getPathExtension(this.props.path);
			if (!utils.supportedExtensions.has(ext)) {
				return; // Don't annotate unsupported languages
			}

			addAnnotations(this.props.path, { repoURI, rev, isDelta: this.isDelta || false, isBase }, this.props.blobElement, this.isSplitDiff || false, loggerProps);
		};

		const applyAnnotationsIfResolvedRev = (uri: string, rev?: string, isBase?: boolean) => {
			const resolvedRev = this.state.resolvedRevs[backend.cacheKey(uri, rev)];
			if (resolvedRev && resolvedRev.commitID) {
				apply(uri, resolvedRev.commitID, Boolean(isBase), this.eventLoggerProps());
			}
		};

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				applyAnnotationsIfResolvedRev(this.baseRepoURI, this.baseCommitID, true);
			}
			if (this.headCommitID && this.headRepoURI) {
				applyAnnotationsIfResolvedRev(this.headRepoURI, this.headCommitID, false);
			}
		} else {
			applyAnnotationsIfResolvedRev(this.props.repoURI, this.rev, false);
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
		if (typeof this.state.resolvedRevs[this.props.repoURI] === "undefined") {
			return null;
		}

		if (!utils.supportedExtensions.has(utils.getPathExtension(this.props.path))) {
			let ariaLabel = !utils.upcomingExtensions.has(utils.getPathExtension(this.props.path)) ? "File not supported" : "Language support coming soon!";

			return (<div style={Object.assign({ cursor: "not-allowed" }, buttonStyle)} className={className} aria-label={ariaLabel}>
				<SourcegraphIcon style={iconStyle} />
				Sourcegraph
			</div>);

		} else if (github.isPrivateRepo() && this.state.resolvedRevs[this.props.repoURI].notFound) {
			// Not signed in or not auth'd for private repos
			return (<div style={buttonStyle} className={className} aria-label={`Authorize Sourcegraph`}>
				<a href={`https://sourcegraph.com/authext?rtg=${encodeURIComponent(window.location.href)}`}
					style={{ textDecoration: "none", color: "inherit" }}>
					<SourcegraphIcon style={Object.assign({ WebkitFilter: "grayscale(100%)" }, iconStyle)} />
					Sourcegraph
				</a>
			</div>);

		} else if (this.state.resolvedRevs[this.props.repoURI].cloneInProgress) {
			return (<div style={buttonStyle} className={className} aria-label={`Sourcegraph is analyzing ${this.props.repoURI.split("github.com/")[1]}`}>
				<SourcegraphIcon style={iconStyle} />
				Loading...
			</div>);

		} else {
			return (<div style={buttonStyle} className={className} aria-label="View on Sourcegraph">
				<a href={this.getBlobUrl()} style={{ textDecoration: "none", color: "inherit" }}><SourcegraphIcon style={iconStyle} />
					Sourcegraph
				</a>
			</div>);
		}
	}
}
