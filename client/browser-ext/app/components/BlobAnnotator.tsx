import * as backend from "../backend";
import * as utils from "../utils";
import { addAnnotations } from "../utils/annotations";
import * as github from "../utils/github";
import { SourcegraphIcon } from "./Icons";
import * as React from "react";

const isCloning = new Set<string>();

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

	constructor(props: Props) {
		super(props);
		this.state = {
			resolvedRevs: {},
		};

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

	_updateResolvedRevs(repo: string, rev?: string): void {
		const key = backend.cacheKey(repo, rev);
		if (this.state.resolvedRevs[key] && this.state.resolvedRevs[key].commitID) {
			return; // nothing to do
		}
		const p = backend.resolveRev(repo, rev);
		if (!p || !p.then) {
			console.error("WHY THE FUCK IS THIS HAPPENING", p, p.then);
			return;
		}
		p.then((resp) => {
			this.setState({ resolvedRevs: Object.assign({}, this.state.resolvedRevs, { [key]: resp }, { [repo]: resp }) });
		});
	}

	resolveRevs(): void {
		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				this._updateResolvedRevs(this.baseRepoURI, this.baseCommitID);
			}
			if (this.headCommitID && this.headRepoURI) {
				this._updateResolvedRevs(this.headRepoURI, this.headCommitID);
			}
		} else if (this.rev) {
			this._updateResolvedRevs(this.props.repoURI, this.rev);
		} else {
			console.error("unable to fetch annotations; missing revision data");
		}
	}

	_addAnnotations(): void {
		const apply = (repoURI: string, rev: string, isBase: boolean, loggerProps: Object) => {
			const ext = utils.getPathExtension(this.props.path);
			if (!utils.supportedExtensions.has(ext)) {
				return; // Don't annotate unsupported languages
			}

			addAnnotations(this.props.path, { repoURI, rev, isDelta: this.isDelta || false, isBase }, this.props.blobElement, this.isSplitDiff || false, loggerProps);
		};

		if (this.isDelta) {
			if (this.baseCommitID && this.baseRepoURI) {
				apply(this.baseRepoURI, this.baseCommitID, true, this.eventLoggerProps());
			}
			if (this.headCommitID && this.headRepoURI) {
				apply(this.headRepoURI, this.headCommitID, false, this.eventLoggerProps());
			}
		} else {
			const resolvedRev = this.state.resolvedRevs[backend.cacheKey(this.props.repoURI, this.rev)];
			if (resolvedRev && resolvedRev.commitID) {
				apply(this.props.repoURI, resolvedRev.commitID, false, this.eventLoggerProps());
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
		if (typeof this.state.resolvedRevs[this.props.repoURI] === "undefined") {
			return null;
		}

		if (github.isPrivateRepo() && this.state.resolvedRevs[this.props.repoURI].authRequired) {
			// Not signed in or not auth'd for private repos
			return (<div style={buttonStyle} className={className} aria-label={`Authorize Sourcegraph for private repos`}>
				<a href={`https://sourcegraph.com/authext?rtg=${encodeURIComponent(window.location.href)}`}
					style={{ textDecoration: "none", color: "inherit" }}>
					<SourcegraphIcon style={Object.assign({ WebkitFilter: "grayscale(100%)" }, iconStyle)} />
					Sourcegraph
				</a>
			</div>);

		} else if (this.state.resolvedRevs[this.props.repoURI].cloneInProgress) {
			// Cloning the repo
			if (!isCloning.has(this.props.repoURI)) {
				isCloning.add(this.props.repoURI);
				this.refreshInterval = setInterval(this.resolveRevs, 5000);
			}

			return (<div style={buttonStyle} className={className} aria-label={`Sourcegraph is analyzing ${this.props.repoURI.split("github.com/")[1]}`}>
				<SourcegraphIcon style={iconStyle} />
				Loading...
			</div>);

		} else if (!utils.supportedExtensions.has(utils.getPathExtension(this.props.path))) {
			let ariaLabel: string;
			if (!utils.upcomingExtensions.has(utils.getPathExtension(this.props.path))) {
				ariaLabel = "File not supported";
			} else {
				ariaLabel = "Language support coming soon!";
			}

			return (<div style={Object.assign({ cursor: "not-allowed" }, buttonStyle)} className={className} aria-label={ariaLabel}>
				<SourcegraphIcon style={iconStyle} />
				Sourcegraph
			</div>);

		} else {
			if (isCloning.has(this.props.repoURI)) {
				isCloning.delete(this.props.repoURI);
				if (this.refreshInterval) {
					clearInterval(this.refreshInterval);
				}
			}

			return (<div style={buttonStyle} className={className} aria-label="View on Sourcegraph">
				<a href={this.getBlobUrl()} style={{ textDecoration: "none", color: "inherit" }}><SourcegraphIcon style={iconStyle} />
					Sourcegraph
				</a>
			</div>);
		}
	}
}
