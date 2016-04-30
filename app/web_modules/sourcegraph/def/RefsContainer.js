// @flow weak

import React from "react";
import update from "react/lib/update";
import CSSModules from "react-css-modules";

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
import {urlToDef, urlToDef2} from "sourcegraph/def/routes";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import styles from "./styles/Refs.css";
import {FileIcon} from "sourcegraph/components/Icons";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {RepoLink, Label} from "sourcegraph/components";
import breadcrumb from "sourcegraph/util/breadcrumb";

const SNIPPET_REF_CONTEXT_LINES = 4; // Number of additional lines to show above/below a ref

class RefsContainer extends Container {
	static propTypes = {
		refRepo: React.PropTypes.string.isRequired,
		initNumSnippets: React.PropTypes.number, // number of snippets to initially expand
		fileCollapseThreshold: React.PropTypes.number, // number of files to show before "and X more..."-style paginator
	};
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.rangesMemo = {}; // optimization: cache the line range that should be displayed for each ref

		// optimization: these memos reduce the amount of component state which must be copied in reconcileState
		this.filesByName = {};
		this.entrySpecsByName = {};
		this.ranges = {};
		this.anns = {};
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
		if (typeof state.showAllFiles === "undefined") {
			state.showAllFiles = false;
		}
		state.fileCollapseThreshold = props.fileCollapseThreshold || 3;

		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.activeDef = state.def ? urlToDef2(state.repo, state.rev, state.def) : state.def;

		state.refLocations = state.def ? DefStore.refLocations.get(state.repo, state.rev, state.def) : null;

		state.refRepo = props.refRepo;
		if (state.refLocations && !state.refCount) {
			state.refCount = state.refLocations.find((loc) => loc.Repo === state.refRepo).Count;
		}
		if (state.refLocations && !state.fileLocations) {
			state.fileLocations = state.refLocations
				.filter((loc) => loc.Repo === state.refRepo)
				.map((loc) => loc.Files);
			state.fileLocations = [].concat.apply([], state.fileLocations).sort((a, b) => b.Count - a.Count); // flatten
		}
		if (state.fileLocations && !state.shownFiles) {
			// Initially show the first three files only
			state.shownFiles = state.fileLocations.map((file, i) => i < (props.initNumSnippets || 0));
			state.shownFileIndex = {}; // keep an index, for lookup by name
			state.fileLocations.forEach((file, i) => state.shownFileIndex[file.Path] = i);
		}

		state.refs = props.refs || DefStore.refs.get(state.repo, state.rev, state.def, state.refRepo, null);

		if (state.mouseover) {
			state.highlightedDef = DefStore.highlightedDef;
			if (state.highlightedDef) {
				let {repo, rev, def} = defRouteParams(state.highlightedDef);
				state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
			} else {
				state.highlightedDefObj = null;
			}
		}

		state.forceComponentUpdate = false;
		if (state.refs && !state.refs.Error) {
			for (let ref of state.refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.rev : ref.CommitID;
				this.entrySpecsByName[ref.File] = this.entrySpecsByName[ref.File] ? this.entrySpecsByName[ref.File] : {RepoRev: {URI: ref.Repo, Rev: refRev}, Path: ref.File};

				if (!this.filesByName[ref.File]) {
					let file = BlobStore.files.get(ref.Repo, refRev, ref.File);
					if (file && !this.filesByName[ref.File]) state.forceComponentUpdate = true;
					this.filesByName[ref.File] = file;
				}

				if (this.filesByName[ref.File]) {
					this.ranges[ref.File] = this.ranges[ref.File] ? this.ranges[ref.File] : [];
					if (this.rangesMemo[ref.File]) {
						this.ranges[ref.File] = this.rangesMemo[ref.File];
					} else {
						let contents = this.filesByName[ref.File].ContentsString;
						this.ranges[ref.File].push([
							Math.max(lineFromByte(contents, ref.Start) - SNIPPET_REF_CONTEXT_LINES, 0),
							lineFromByte(contents, ref.End) + SNIPPET_REF_CONTEXT_LINES,
						]);
						this.rangesMemo[ref.File] = this.ranges[ref.File];
					}
				}
				if (!this.anns[ref.File]) {
					let anns = BlobStore.annotations.get(ref.Repo, refRev, ref.CommitID, ref.File);
					if (anns && !this.anns[ref.File]) state.forceComponentUpdate = true;
					this.anns[ref.File] = anns;
				}
			}
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations(nextState.repo, nextState.rev, nextState.def));
		}

		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.refRepo !== nextState.refRepo) {
			if (nextState.refRepo) {
				Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.refRepo));
			}
		}

		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			const {repo, rev, def} = defRouteParams(nextState.highlightedDef);
			Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
		}

		if (nextState.refs && !nextState.refs.Error && (nextState.refs !== prevState.refs || nextState.shownFiles !== prevState.shownFiles)) {
			for (let ref of nextState.refs) {
				let refRev = ref.Repo === nextState.repo ? nextState.rev : ref.CommitID;
				if (nextState.shownFiles && nextState.shownFiles[nextState.shownFileIndex[ref.File]]) {
					Dispatcher.Backends.dispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
					Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(ref.Repo, refRev, ref.CommitID, ref.File));
				}
			}
		}
	}

	renderFileHeader(entrySpec, count, i) {
		let pathBreadcrumb = breadcrumb(
			entrySpec.Path,
			(j) => <span key={j} className={styles.sep}>/</span>,
			(path, component, j, isLast) => <span className={styles.pathPart} key={j}>{component}</span>
		);
		return (
			<h3 key={entrySpec.Path} className={styles.filename}>
				<Label outline={true} style={{display: "inline-block", marginRight: "5px"}}>{`${count} ref${count > 1 ? "s" : ""}`}</Label>
				<a href={urlToBlob(entrySpec.RepoRev.URI, entrySpec.RepoRev.Rev, entrySpec.Path)}
					onClick={(e) => {
						e.preventDefault();
						this.setState(update(this.state, {shownFiles: {$splice: [[i, 1, !this.state.shownFiles[i]]]}}));
					}}>{pathBreadcrumb}</a>
				<Link to={urlToBlob(entrySpec.RepoRev.URI, entrySpec.RepoRev.Rev, entrySpec.Path)}><span className={styles.fileIcon}><FileIcon /></span></Link>
			</h3>
		);
	}

	paginatorText() {
		const remainder = this.state.fileLocations.slice(this.state.fileCollapseThreshold);
		const count = remainder.reduce((memo, file) => memo + file.Count, 0);
		return `${count} more ref${count > 1 ? "s" : ""} in ${remainder.length} file${remainder.length > 1 ? "s" : ""}`;
	}

	render() {
		if (this.state.refLocations && this.state.refLocations.Error) {
			return (
				<Header
					title={`${httpStatusCode(this.state.refLocations.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		let def = this.state.defObj;

		return (
			<div className={styles.container}
				onMouseOver={() => this.setState({mouseover: true})}
				onMouseOut={() => this.setState({mouseover: false})}>
				{/* mouseover state is for optimization which will only re-render the moused-over blob when a def is highlighted */}
				{/* this is important since there may be many ref containers on the page */}
				<div>
					{this.state.refCount &&
						<h2 className={styles.repo}>
							<Label outline={true} style={{display: "inline-block", marginRight: "10px", fontSize: "18px"}}>
								{`${this.state.refCount} refs`}
							</Label>
							<RepoLink repo={this.state.refRepo} />
						</h2>
					}
					<hr/>
					<div className={styles.refs}>
						{this.state.fileLocations && this.state.fileLocations.map((loc, i) => {
							if (!this.entrySpecsByName[loc.Path] || (!this.state.showAllFiles && i >= this.state.fileCollapseThreshold)) return null;

							let entrySpec = this.entrySpecsByName[loc.Path];
							if (!this.state.shownFiles[i]) return this.renderFileHeader(entrySpec, loc.Count, i);

							let file = this.filesByName ? this.filesByName[loc.Path] : null;
							if (!file) {
								return <div key={i}>{this.renderFileHeader(entrySpec, loc.Count, i)}<BlobContentPlaceholder key={i} numLines={10} /></div>;
							}
							let path = entrySpec.Path;
							let repoRev = entrySpec.RepoRev;
							return (
								<div key={i}>
									{this.renderFileHeader(entrySpec, loc.Count, i)}
									<Blob
										repo={this.state.refRepo}
										rev={repoRev.Rev}
										path={path}
										contents={file.ContentsString}
										annotations={this.anns[path] || null}
										activeDef={this.state.activeDef}
										activeDefNoRev={this.state.activeDef ? urlToDef(def, "") : null}
										lineNumbers={true}
										displayRanges={this.ranges[path] || null}
										highlightedDef={this.state.highlightedDef || null}
										highlightedDefObj={this.state.highlightedDefObj || null} />
								</div>
							);
						})}
					</div>
					{!this.state.showAllFiles && this.state.fileLocations && this.state.fileLocations.length > this.state.fileCollapseThreshold &&
						<div className={styles.paginator}>
							<span className={styles.pageLink} onClick={() => this.setState({showAllFiles: true})}>{this.paginatorText()}</span>
						</div>
					}
				</div>
				{this.state.highlightedDefObj && !this.state.highlightedDefObj.Error && <DefTooltip currentRepo={this.state.repo} def={this.state.highlightedDefObj} />}
			</div>
		);
	}
}

export default CSSModules(RefsContainer, styles);
