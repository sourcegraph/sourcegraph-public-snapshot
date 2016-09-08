// tslint:disable: typedef ordered-imports

import {Location} from "history";
import * as React from "react";
import * as classNames from "classnames";

import {BlobLegacy} from "sourcegraph/blob/BlobLegacy";

import {BlobStore, keyForFile, keyForAnns} from "sourcegraph/blob/BlobStore";
import {BlobContentPlaceholder} from "sourcegraph/blob/BlobContentPlaceholder";
import {Container} from "sourcegraph/Container";
import {Store} from "sourcegraph/Store";
import {DefStore} from "sourcegraph/def/DefStore";
import {DefTooltip} from "sourcegraph/def/DefTooltip";
import {Link} from "react-router";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/blob/BlobBackend";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {lineFromByte} from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import {Header} from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import {RepoLink} from "sourcegraph/components";
import {FaAngleDown, FaAngleRight} from "sourcegraph/components/Icons";
import {breadcrumb} from "sourcegraph/util/breadcrumb";
import {stripDomain} from "sourcegraph/util/stripDomain";
import * as styles from "sourcegraph/def/styles/Refs.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {FaThumbsUp, FaThumbsDown} from "sourcegraph/components/Icons";
import {context} from "sourcegraph/app/context";

const SNIPPET_REF_CONTEXT_LINES = 4; // Number of additional lines to show above/below a ref

interface Props {
	location: Location;
	repo?: string;
	rev?: string;
	commitID?: string;
	def?: string;
	defObj?: any;
	refs?: any;
	repoRefs: {
		Repo?: string;
		Files: any[];
	};
	refetch?: boolean;
	initNumSnippets?: number; // number of snippets to initially expand
	fileCollapseThreshold?: number; // number of files to show before "and X more..."-style paginator
	rangeLimit?: number;
	showRepoTitle?: boolean;
	refIndex?: number;
}

type State = any;

export class RefsContainer extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	rangesMemo: any;
	filesByName: any;
	ranges: any;
	anns: any;

	constructor(props: Props) {
		super(props);
		this.state = {
			shownFiles: new Set(),
			initExpanded: false, // Keep track of when we've auto-expanded snippets.
		};
		this.rangesMemo = {}; // optimization: cache the line range that should be displayed for each ref

		// optimization: these memos reduce the amount of component state which must be copied in reconcileState
		this.filesByName = {};
		this.ranges = {};
		this.anns = {};
		this._toggleFile = this._toggleFile.bind(this);
		this._vote = this._vote.bind(this);
	}

	shouldComponentUpdate(nextProps, nextState, nextContext) {
		if (super.shouldComponentUpdate(nextProps, nextState, nextContext)) {
			return true;
		}

		// Since the reference values of the memo'd state don't change even though the contents
		// may be updated (e.g. as a result of asynchronous fetches) reconcileState
		// must set a special flag if it resolves new data from a store which is kept in the memo.
		return Boolean(nextState.forceComponentUpdate);
	}

	stores(): Store<any>[] {
		return [DefStore, BlobStore];
	}

	reconcileState(state: State, props: Props): void {
		state.location = props.location || null;
		if (typeof state.showAllFiles === "undefined") {
			state.showAllFiles = false;
		}
		state.fileCollapseThreshold = props.fileCollapseThreshold || 3;

		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.showRepoTitle = props.showRepoTitle || false;

		state.refRepo = props.repoRefs.Repo || null;
		state.refRev = state.refRepo === state.repo ? state.rev : null;
		state.repoRefLocations = props.repoRefs || null;
		state.rangeLimit = props.rangeLimit || null;
		if (state.repoRefLocations) {
			state.fileLocations = state.repoRefLocations.Files;
		}

		state.refs = props.refs || DefStore.refs.get(state.repo, state.rev, state.def, state.refRepo, null);

		state.hoverInfos = DefStore.hoverInfos;
		state.hoverPos = DefStore.hoverPos;

		if (state.fileLocations && !state.initExpanded) {
			// Auto-expand N snippets by default.
			for (let i = 0; i < props.initNumSnippets; i++) {
				let loc = state.fileLocations[i];
				if (loc) {
					state.shownFiles.add(loc.Path);
				}
			}
			state.initExpanded = true;
		}

		state.forceComponentUpdate = false;
		if (state.refs && !state.refs.Error) {
			for (let ref of state.refs || []) {
				if (!ref) {
					continue;
				}
				let refRev = ref.CommitID;
				if (!this.filesByName[ref.File]) {
					let file = BlobStore.files[keyForFile(ref.Repo, refRev, ref.File)] || null;
					if (file) {
						// Pass through Error to this.filesByName (i.e., proceed even if file.Error is truthy).
						state.forceComponentUpdate = true;
						this.filesByName[ref.File] = file;
					}
				}

				if (this.filesByName[ref.File] && !this.filesByName[ref.File].Error) {
					this.ranges[ref.File] = this.ranges[ref.File] ? this.ranges[ref.File] : [];
					const rangeKey = `${ref.File}${ref.Start}`;
					if (!this.rangesMemo[rangeKey]) {
						// Response from LSP comes with StartLine, so we can directly use it if available.
						let startLineNum;
						let endLineNum;
						if (ref.StartLine) {
							startLineNum = ref.StartLine;
							endLineNum = ref.EndLine;
						} else {
							let contents = this.filesByName[ref.File].ContentsString;
							startLineNum = lineFromByte(contents, ref.Start);
							endLineNum = lineFromByte(contents, ref.End);
						}
						this.ranges[ref.File].push([
							Math.max(startLineNum - SNIPPET_REF_CONTEXT_LINES, 0),
							endLineNum + SNIPPET_REF_CONTEXT_LINES,
							startLineNum,
						]);
						this.rangesMemo[rangeKey] = true;
					}
				}
				if (!this.anns[ref.File]) {
					let anns = BlobStore.annotations[keyForAnns(ref.Repo, ref.CommitID, ref.File)] || null;
					if (anns) {
						// Pass through Error to this.anns (i.e., proceed even if anns.Error is truthy).
						state.forceComponentUpdate = true;
						this.anns[ref.File] = anns;
					}
				}
			}
		}
	}

	onStateTransition(prevState: State, nextState: State): void {
		const refPropsUpdated = prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.refRepo !== nextState.refRepo;
		if (refPropsUpdated) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.refRepo));
		}

		if (nextState.refs && nextState.refs.length > 0 && !nextState.refs.Error && (nextState.refs !== prevState.refs || nextState.shownFiles !== prevState.shownFiles)) {
			let firstRef = nextState.refs[0]; // hack: assuming that all refs given to a RefsContainer are from the same repo and rev, thus using the first ref to determine which files we want to show
			let repo = firstRef.Repo;
			let rev = firstRef.CommitID;
			for (let file of Array.from(nextState.shownFiles as Set<any>)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(repo, rev, file));
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(repo, rev, file));
			}
		}
	}

	renderFileHeader(repo, rev, path, count, i) {
		let trimmedPath = stripDomain(repo);
		trimmedPath = trimmedPath.concat("/", path);
		let pathBreadcrumb = breadcrumb(
			trimmedPath,
			(j) => <span key={j}> / </span>,
			(_, component, j, isLast) => {
				let span = <span key={j}>{component}</span>;
				if (isLast) {
					return <Link className={styles.pathEnd} to={urlToBlob(repo, rev, path)} key={j}> {span} </Link>;
				}
				return span;
			}
		);
		return (
			<div key={path} className={styles.filename} onClick={(e) => {
				if (e.button !== 0) {
					return; // only expand on main button click
				}
				this._toggleFile(path);
			}}>
				<div className={styles.breadcrumbIcon}>
					{this.state.shownFiles.has(path) ? <FaAngleDown className={styles.toggleIcon} /> : <FaAngleRight className={styles.toggleIcon} />}
				</div>
				<div className={styles.pathContainer}>
					{pathBreadcrumb}
					{count &&
						<span className={styles.refsLabel}>{`${count} ref${count > 1 ? "s" : ""}`}</span>
					}
				</div>
			</div>
		);
	}

	paginatorText() {
		const remainder = this.state.fileLocations.slice(this.state.fileCollapseThreshold);
		const count = remainder.reduce((memo, file) => memo + file.Count, 0);
		return `Used ${count} more time${count > 1 ? "s" : ""} in ${remainder.length} other file${remainder.length > 1 ? "s" : ""} ...`;
	}

	_toggleFile(path) {
		let newOpenFiles = new Set();
		this.state.shownFiles.forEach((f) => newOpenFiles.add(f));
		if (this.state.shownFiles.has(path)) {
			newOpenFiles.delete(path);
		} else {
			newOpenFiles.add(path);
			(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REFERENCES, AnalyticsConstants.ACTION_TOGGLE, "RefsFileExpanded", {repo: this.props.repo, def: this.props.def});
		}
		this.setState({shownFiles: newOpenFiles});
	}

	_vote(upvote, repo, path) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_INTERNAL, AnalyticsConstants.ACTION_CLICK, "UsageExampleVote", {
			page: window.location.href,
			upvote: upvote,
			repo: repo,
			path: path,
			index: this.props.refIndex,
		});
		this.setState({voteDone: true});
	}

	_clickedFromRepo() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "FromRepoClicked", {repo: this.state.repo, def: this.state.def, ref_repo: this.state.refRepo});
	}

	render(): JSX.Element | null {
		if (this.state.fileLocations && this.state.fileLocations.length === 0) {
			return null;
		}

		if (this.state.refLocations && this.state.refLocations.Error) {
			return (
				<Header
					title={`${httpStatusCode(this.state.refLocations.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		if (this.state.refs) {
			if (this.state.refs.Error) {
				console.error("Error fetching refs", this.state.refs.Error.response);
				return null;
			}
			for (let loc of this.state.fileLocations) {
				// Do not display a file without refs.
				if (this.state.refs.filter((r) => r.File === loc.Path).length === 0) {
					console.error(`No references for ${this.state.def} found in ${this.state.refRepo}/${loc.Path}`);
					return null;
				}
			}
		}

		return (
			<div className={classNames(base.pa4, base.bb, colors.b__cool_pale_gray, styles.full_width_sm)}>
				{this.state.showRepoTitle &&
					<div>
						<RepoLink className={styles.repoLink} repo={this.state.refRepo} />
					</div>
				}
				<div className={styles.container}
					onMouseEnter={() => {
						if (!this.state.mouseover) {
							this.setState({mouseover: true, mouseout: false});
						}
					}}
					onMouseLeave={() => this.setState({mouseover: false, mouseout: true})}
					onMouseOut={() => Dispatcher.Stores.dispatch(new DefActions.Hovering(null))}>
					{/* mouseover state is for optimization which will only re-render the moused-over blob when a def is highlighted */}
					{/* this is important since there may be many ref containers on the page */}
					<div>
						<div>
							{this.state.fileLocations && this.state.fileLocations.map((loc, i) => {
								if (!this.state.showAllFiles && i >= this.state.fileCollapseThreshold) {
									return null;
								}

								let err;
								let file = this.filesByName ? this.filesByName[loc.Path] : null;
								if (file && file.Error) {
									switch (file.Error.response.status) {
									case 413:
										err = "Sorry, this file is too large to display.";
										break;
									default:
										err = "File is not available.";
									}
								}
								if (this.state.refs && this.state.refs.Error) {
									err = `Error loading references for ${loc.Path}.`;
								}

								if (err) {
									return <div key={i}>{this.renderFileHeader(this.state.refRepo, this.state.refRev, loc.Path, loc.Count, i)}<p className={styles.fileError}>{err}</p></div>;
								}
								if (!file) {
									return <div key={i}><BlobContentPlaceholder key={i} numLines={SNIPPET_REF_CONTEXT_LINES * 2 + 1} /></div>;
								}

								let ranges = this.ranges[loc.Path];
								if (this.state.rangeLimit) {
									ranges = ranges.slice(0, this.state.rangeLimit);
									ranges.map((r) => [r[0], Math.min(r[0] + 10, r[1])]);
								}

								let voteStyle = this.state.voteDone ? styles.voteDone : styles.vote;
								return (
									<div key={i} className={styles.single_ref_container}>
										{context.user && context.user.Admin && <div className={classNames(voteStyle, styles.left_align_sm)}>
											<a className={styles.upvote} onClick={() => this._vote(true, this.state.refRepo, loc.Path)}><FaThumbsUp /></a>
											<a className={styles.downvote} onClick={() => this._vote(false, this.state.refRepo, loc.Path)}><FaThumbsDown /></a>
										</div>}
										<div className={styles.refs}>
											<BlobLegacy
												location={this.state.location}
												repo={this.state.refRepo}
												rev={this.state.refRev}
												commitID={this.state.commitID}
												path={loc.Path}
												contents={file.ContentsString}
												annotations={this.anns[loc.Path] || null}
												skipAnns={file.ContentsString && file.ContentsString.length >= 40 * 2500}
												activeDefRepo={this.state.repo}
												activeDef={this.state.def}
												lineNumbers={false}
												displayRanges={ranges || null}
												highlightedDef={null}
												highlightedDefObj={null}
												textSize="large" />
										</div>
										{this.state.refRepo && <div className={classNames(base.mt3, styles.f7, base.hidden_s)}>From <Link to={`${urlToBlob(this.state.refRepo, this.state.refRev, loc.Path)}${ranges ? `#L${ranges[0][2]}` : ""}`} onClick={this._clickedFromRepo.bind(this)}>{this.state.refRepo}</Link></div>}
									</div>
								);
							})}
						</div>

						{!this.state.showAllFiles && this.state.fileLocations && this.state.fileLocations.length > this.state.fileCollapseThreshold &&
							<div className={styles.filename} onClick={() => this.setState({showAllFiles: true})}>
								<FaAngleRight className={styles.toggleIcon} />
								{this.paginatorText()}
							</div>
						}
					</div>
					<DefTooltip
						currentRepo={this.state.repo}
						hoverPos={this.state.hoverPos}
						hoverInfos={this.state.hoverInfos} />
				</div>
			</div>
		);
	}
}
