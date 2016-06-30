// @flow weak

import React from "react";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import BlobContentPlaceholder from "sourcegraph/blob/BlobContentPlaceholder";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import {Link} from "react-router";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/blob/BlobBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {routeParams as defRouteParams} from "sourcegraph/def";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {RepoLink} from "sourcegraph/components";
import {FaAngleDown, FaAngleRight} from "sourcegraph/components/Icons";
import breadcrumb from "sourcegraph/util/breadcrumb";
import stripDomain from "sourcegraph/util/stripDomain";
import styles from "./styles/Refs.css";
import base from "sourcegraph/components/styles/_base.css";
import colors from "sourcegraph/components/styles/_colors.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const SNIPPET_REF_CONTEXT_LINES = 0; // Number of additional lines to show above/below a ref

export default class RefsContainer extends Container {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string.isRequired,
		def: React.PropTypes.string.isRequired,
		defObj: React.PropTypes.object.isRequired,
		repoRefs: React.PropTypes.shape({
			Repo: React.PropTypes.string,
			Files: React.PropTypes.array,
		}),
		refetch: React.PropTypes.bool,
		initNumSnippets: React.PropTypes.number, // number of snippets to initially expand
		fileCollapseThreshold: React.PropTypes.number, // number of files to show before "and X more..."-style paginator
		rangeLimit: React.PropTypes.number,
		showRepoTitle: React.PropTypes.bool,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
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
	}

	shouldComponentUpdate(nextProps, nextState, nextContext) {
		if (super.shouldComponentUpdate(nextProps, nextState, nextContext)) return true;

		// Since the reference values of the memo'd state don't change even though the contents
		// may be updated (e.g. as a result of asynchronous fetches) reconcileState
		// must set a special flag if it resolves new data from a store which is kept in the memo.
		return Boolean(nextState.forceComponentUpdate);
	}


	stores() {
		return [DefStore, BlobStore];
	}

	reconcileState(state, props) {
		state.prefetch = props.prefetch || false;

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

		if (state.fileLocations && !state.initExpanded) {
			// Auto-expand N snippets by default.
			for (let i=0; i<props.initNumSnippets; i++) {
				let loc = state.fileLocations[i];
				if (loc) state.shownFiles.add(loc.Path);
			}
			state.initExpanded = true;
		}

		if (state.mouseover) {
			state.highlightedDef = DefStore.highlightedDef;
			if (state.highlightedDef) {
				let {repo, rev, def} = defRouteParams(state.highlightedDef);
				state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
			} else {
				state.highlightedDefObj = null;
			}
		}
		if (state.mouseout) {
			// Clear DefTooltip so it doesn't hang around.
			state.highlightedDef = null;
			state.highlightedDefObj = null;
		}

		state.forceComponentUpdate = false;
		if (state.refs && !state.refs.Error) {
			for (let ref of state.refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.commitID : ref.CommitID;
				if (!this.filesByName[ref.File]) {
					let file = BlobStore.files.get(ref.Repo, refRev, ref.File);
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
						let contents = this.filesByName[ref.File].ContentsString;
						this.ranges[ref.File].push([
							Math.max(lineFromByte(contents, ref.Start) - SNIPPET_REF_CONTEXT_LINES, 0),
							lineFromByte(contents, ref.End) + SNIPPET_REF_CONTEXT_LINES,
						]);
						this.rangesMemo[rangeKey] = true;
					}
				}
				if (!this.anns[ref.File]) {
					let anns = BlobStore.annotations.get(ref.Repo, ref.CommitID, ref.File);
					if (anns) {
						// Pass through Error to this.anns (i.e., proceed even if anns.Error is truthy).
						state.forceComponentUpdate = true;
						this.anns[ref.File] = anns;
					}
				}
			}
		}
	}

	onStateTransition(prevState, nextState) {
		const refPropsUpdated = prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.refRepo !== nextState.refRepo;
		if (refPropsUpdated) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.refRepo));
		}

		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			const {repo, rev, def} = defRouteParams(nextState.highlightedDef);
			Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
		}

		if (nextState.refs && !nextState.refs.Error && (nextState.refs !== prevState.refs || nextState.shownFiles !== prevState.shownFiles)) {
			for (let ref of nextState.refs) {
				let refRev = ref.Repo === nextState.repo ? nextState.commitID : ref.CommitID;
				if (nextState.shownFiles.has(ref.File)) {
					Dispatcher.Backends.dispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
					Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(ref.Repo, ref.CommitID, ref.File));
				}
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
				if (e.button !== 0) return; // only expand on main button click
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
			this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REFERENCES, AnalyticsConstants.ACTION_TOGGLE, "RefsFileExpanded", {repo: this.props.repo, def: this.props.def});
		}
		this.setState({shownFiles: newOpenFiles});
	}

	render() {
		if (this.state.fileLocations && this.state.fileLocations.length === 0) return null;

		if (this.state.refLocations && this.state.refLocations.Error) {
			return (
				<Header
					title={`${httpStatusCode(this.state.refLocations.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		return (
			<div className={`${base.pa4} ${base.bb} ${colors["b--cool-pale-gray"]}`}>
			{this.state.showRepoTitle &&
				<div>
					<RepoLink className={styles.repoLink} repo={this.state.refRepo} />
				</div>
			}
				<div className={styles.container}
					onMouseEnter={() => {
						if (!this.state.mouseover) this.setState({mouseover: true, mouseout: false});
					}}
					onMouseLeave={() => this.setState({mouseover: false, mouseout: true})}
					onMouseOut={() => Dispatcher.Stores.dispatch(new DefActions.HighlightDef(null))}>
					{/* mouseover state is for optimization which will only re-render the moused-over blob when a def is highlighted */}
					{/* this is important since there may be many ref containers on the page */}
					<div>
						<div className={styles.refs}>
							{this.state.fileLocations && this.state.fileLocations.map((loc, i) => {
								if (!this.state.showAllFiles && i >= this.state.fileCollapseThreshold) return null;
								if (!this.state.shownFiles.has(loc.Path)) return this.renderFileHeader(this.state.refRepo, this.state.refRev, loc.Path, loc.Count, i);

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
								} else if (this.state.refs && this.state.refs.filter((r) => r.File === loc.Path).length === 0) {
									err = `No references found for ${loc.Path}`;
								}

								if (err) {
									return <div key={i}>{this.renderFileHeader(this.state.refRepo, this.state.refRev, loc.Path, loc.Count, i)}<p className={styles.fileError}>{err}</p></div>;
								}
								if (!file) {
									return <div key={i}>{this.renderFileHeader(this.state.refRepo, this.state.refRev, loc.Path, loc.Count, i)}<BlobContentPlaceholder key={i} numLines={10} /></div>;
								}

								let ranges = this.ranges[loc.Path];
								if (this.state.rangeLimit) {
									ranges = ranges.slice(0, this.state.rangeLimit);
									ranges.map((r) => [r[0], Math.min(r[0] + 10, r[1])]);
								}

								return (
									<div key={i}>
										<Blob
											repo={this.state.refRepo}
											rev={this.state.refRev}
											path={loc.Path}
											contents={file.ContentsString}
											annotations={this.anns[loc.Path] || null}
											skipAnns={file.ContentsString && file.ContentsString.length >= 40*2500}
											activeDefRepo={this.state.repo}
											activeDef={this.state.def}
											lineNumbers={false}
											displayRanges={ranges || null}
											highlightedDef={this.state.highlightedDef || null}
											highlightedDefObj={this.state.highlightedDefObj || null}
											textSize="large"
											className={styles.blob} />
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
					{this.state.highlightedDefObj && !this.state.highlightedDefObj.Error && <DefTooltip currentRepo={this.state.repo} def={this.state.highlightedDefObj} />}
				</div>
				{this.state.refRepo && <div className={`${base.mt3} ${styles.f7}`}>Used in <Link to={this.state.refRepo}>{this.state.refRepo}</Link></div>}
			</div>
		);
	}
}
