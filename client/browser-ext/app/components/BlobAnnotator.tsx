import * as React from "react";
import * as backend from "../backend";
import * as utils from "../utils";
import { addAnnotations, RepoRevSpec } from "../utils/annotations";
import { eventLogger, sourcegraphUrl } from "../utils/context";
import * as github from "../utils/github";
import { CodeCell, GitHubBlobUrl, GitHubMode, GitHubUrl } from "../utils/types";
import { SourcegraphIcon } from "./Icons";

const className = "btn btn-sm tooltipped tooltipped-n";
const buttonStyle = { marginRight: "5px" };
const iconStyle = { marginTop: "-1px", paddingRight: "4px", fontSize: "18px" };

interface Props {
	headPath: string;
	basePath: string | null;
	repoURI: string;
	fileElement: HTMLElement;
}

interface State {
	resolvedRevs: { [key: string]: backend.ResolvedRevResp };
}

export class BlobAnnotator extends React.Component<Props, State> {
	revisionChecker: NodeJS.Timer;

	// language is determined by the path extension
	fileExtension: string;

	isDelta?: boolean;
	isCommit?: boolean;
	isPullRequest?: boolean;
	isSplitDiff?: boolean;

	// rev is defined for blob view
	rev?: string;

	// base/head properties are defined for diff views (commit + pull request)
	baseCommitID?: string;
	headCommitID?: string;
	baseRepoURI?: string;
	headRepoURI?: string;

	constructor(props: Props) {
		super(props);
		this.state = {
			resolvedRevs: {},
		};

		this.fileExtension = utils.getPathExtension(props.headPath);

		let { isDelta, isPullRequest, isCommit, rev } = utils.parseURL(window.location);
		const gitHubState = github.getGitHubState(window.location.href);
		// TODO(uforic): Eventually, use gitHubState for everything, but for now, only use it when the branch should have a 
		// slash in it to fix that bug
		if (gitHubState && gitHubState.mode === GitHubMode.Blob && (gitHubState as GitHubBlobUrl).rev.indexOf("/") > 0) {
			// correct in case branch has slash in it
			rev = (gitHubState as GitHubBlobUrl).rev;
		}
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

		this.resolveRevs();
		this.addAnnotations();
	}

	componentDidMount(): void {
		this.props.fileElement.addEventListener("click", this.clickRefresh);
		// Set a timer to re-check revision data every 10 seconds, for repos that haven't been
		// cloned and revs that haven't been sync'd to Sourcegraph.com.
		// Single-flighted requests / caching prevents spamming the API.
		this.revisionChecker = setInterval(() => this.resolveRevs(), 5000);
	}

	componentWillUnmount(): void {
		if (this.revisionChecker) {
			clearInterval(this.revisionChecker);
		}
		this.props.fileElement.removeEventListener("click", this.clickRefresh);
	}

	componentDidUpdate(): void {
		// Reapply annotations after reducer state changes.
		this.addAnnotations();
	}

	clickRefresh = (): void => {
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
		}).catch(error => {
			// NO-OP
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

	private getCodeCells(isSplitDiff: boolean, repoRevSpec: RepoRevSpec, el: HTMLElement): CodeCell[] {
		// The blob is represented by a table; the first column is the line number,
		// the second is code. Each row is a line of code
		const table = el.querySelector("table");
		if (!table) {
			return [];
		}
		return github.getCodeCellsForAnnotation(table, Object.assign({ isSplitDiff }, repoRevSpec));
	}

	addAnnotations = (): void => {
		if (!utils.supportedExtensions.has(this.fileExtension)) {
			return; // Don't annotate unsupported languages
		}
		// this check is for either when the blob is collapsed or the dom element is not rendered
		const blobElement = github.tryGetBlobElement(this.props.fileElement);
		if (!blobElement) {
			return;
		}

		/**
		 * applyAnnotationsIfResolvedRev will call addAnnotations if we've established that the repo@rev exists at Sourcegraph
		 */
		const applyAnnotationsIfResolvedRev = (path: string, uri: string, rev?: string, isBase?: boolean) => {
			const resolvedRev = this.state.resolvedRevs[backend.cacheKey(uri, rev)];
			if (resolvedRev && resolvedRev.commitID) {
				const repoRevSpec = { repoURI: uri, rev: resolvedRev.commitID, isDelta: this.isDelta || false, isBase: Boolean(isBase) };
				const cells = this.getCodeCells(this.isSplitDiff || false, repoRevSpec, blobElement);
				addAnnotations(path, repoRevSpec, blobElement, this.getEventLoggerProps(), cells);
			}
		};

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				applyAnnotationsIfResolvedRev(this.props.basePath ? this.props.basePath : this.props.headPath, this.baseRepoURI, this.baseCommitID, true);
			}
			if (this.headCommitID && this.headRepoURI) {
				applyAnnotationsIfResolvedRev(this.props.headPath, this.headRepoURI, this.headCommitID, false);
			}
		} else {
			applyAnnotationsIfResolvedRev(this.props.headPath, this.props.repoURI, this.rev, false);
		}
	}

	getEventLoggerProps(): Object {
		return {
			repo: this.props.repoURI,
			path: this.props.headPath,
			isPullRequest: this.isPullRequest,
			isCommit: this.isCommit,
			language: this.fileExtension,
			isPrivateRepo: github.isPrivateRepo(),
		};
	}

	render(): JSX.Element | null {
		if (!this.isDelta && !Boolean(this.state.resolvedRevs[this.props.repoURI])) {
			return null;
		}
		if (this.isDelta && !Boolean(this.state.resolvedRevs[this.baseRepoURI as string])) {
			return null;
		}
		// this is crappy, and only works because we stick in the cache both the repoURI as key as well as the repoURI@revision
		const resolvedRevs = this.state.resolvedRevs[this.props.repoURI] as backend.ResolvedRevResp;
		return getSourcegraphButton(utils.supportedExtensions.has(this.fileExtension),
			github.isPrivateRepo() && resolvedRevs.notFound as boolean,
			resolvedRevs.cloneInProgress as boolean,
			this.props.repoURI.split("github.com/")[1],
			this.isDelta ? utils.getSourcegraphBlobUrl(sourcegraphUrl, this.headRepoURI as string, this.props.headPath, this.headCommitID) : utils.getSourcegraphBlobUrl(sourcegraphUrl, this.props.repoURI, this.props.headPath, this.rev),
			utils.upcomingExtensions.has(this.fileExtension),
			this.getFileOpenCallback,
			this.getAuthFileCallback);
	}

	getFileOpenCallback = (): void => {
		eventLogger.logOpenFile(this.getEventLoggerProps());
	}

	getAuthFileCallback = (): void => {
		eventLogger.logAuthClicked(this.getEventLoggerProps());
	}
}

function getSourcegraphButton(isFileSupported: boolean, cantFindPrivateRepo: boolean, isLoading: boolean, repoName: string, blobUrl: string, supportedSoon: boolean, fileCallack: () => void, authCallback: () => void): JSX.Element {
	if (!isFileSupported) {
		let ariaLabel = !supportedSoon ? "File not supported" : "Language support coming soon!";
		return (<div style={Object.assign({ cursor: "not-allowed", WebkitFilter: "grayscale(100%)" }, buttonStyle)} className={className} aria-label={ariaLabel}>
			<SourcegraphIcon style={iconStyle} />
			Sourcegraph
		</div>);
	} else if (cantFindPrivateRepo) {
		// Not signed in or not auth'd for private repos
		return (<a href={`${sourcegraphUrl}/login?private=true`}
			style={{ textDecoration: "none", color: "inherit" }} onClick={authCallback}>
			<div style={buttonStyle} className={className} aria-label={`Authorize Sourcegraph`}>
				<SourcegraphIcon style={Object.assign({ WebkitFilter: "grayscale(100%)" }, iconStyle)} />
				Sourcegraph
			</div>
		</a>);
	} else if (isLoading) {
		return (<div style={buttonStyle} className={className} aria-label={`Sourcegraph is analyzing ${repoName}`}>
			<SourcegraphIcon style={iconStyle} />
			Loading...
		</div>);
	}
	return (<a href={blobUrl} style={{ textDecoration: "none", color: "inherit" }} onClick={fileCallack}>
		<div style={buttonStyle} className={className} aria-label="View on Sourcegraph"><SourcegraphIcon style={iconStyle} />
			Sourcegraph
		</div>
	</a>);
}
