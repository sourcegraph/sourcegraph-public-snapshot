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
			state.shownFiles = state.shownFiles ? state.shownFiles : state.fileLocations.map((file, i) => i < (props.initNumSnippets || 0));
		}

		state.refs = props.refs || DefStore.refs.get(state.repo, state.rev, state.def, state.refRepo, null);
		state.filesByName = null;
		state.entrySpecsByName = null;
		state.ranges = null;
		state.anns = null;

		state.highlightedDef = DefStore.highlightedDef;
		if (state.highlightedDef) {
			let {repo, rev, def} = defRouteParams(state.highlightedDef);
			state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
		} else {
			state.highlightedDefObj = null;
		}

		if (state.refs && !state.refs.Error) {
			let filesByName = {};
			let entrySpecsByName = {};
			let ranges = {};
			let anns = {};
			let fileIndex = new Map();
			for (let ref of state.refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.rev : ref.CommitID;
				if (!fileIndex.has(ref.File)) {
					let file = BlobStore.files.get(ref.Repo, refRev, ref.File);
					filesByName[ref.File] = file;
					entrySpecsByName[ref.File] = {RepoRev: {URI: ref.Repo, Rev: refRev}, Path: ref.File};
					ranges[ref.File] = [];
					fileIndex.set(ref.File, file);
				}
				let file = fileIndex.get(ref.File);
				if (file) {
					if (this.rangesMemo[ref.File]) {
						ranges[ref.File] = this.rangesMemo[ref.File];
					} else {
						let contents = file.ContentsString;
						ranges[ref.File].push([
							Math.max(lineFromByte(contents, ref.Start) - SNIPPET_REF_CONTEXT_LINES, 0),
							lineFromByte(contents, ref.End) + SNIPPET_REF_CONTEXT_LINES,
						]);
						this.rangesMemo[ref.File] = ranges[ref.File];
					}
				}
				anns[ref.File] = BlobStore.annotations.get(ref.Repo, refRev, ref.CommitID, ref.File);
			}
			state.filesByName = filesByName;
			state.entrySpecsByName = entrySpecsByName;
			state.ranges = ranges;
			state.anns = anns;
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
			let {repo, rev, def} = defRouteParams(nextState.highlightedDef);
			Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
		}

		if (nextState.refs && !nextState.refs.Error && (nextState.refs !== prevState.refs)) {
			let wantedFiles = new Set();
			for (let ref of nextState.refs) {
				if (wantedFiles.has(ref.File)) continue; // Prevent many requests for the same file.
				let refRev = ref.Repo === nextState.repo ? nextState.rev : ref.CommitID;
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(ref.Repo, refRev, ref.CommitID, ref.File));
				wantedFiles.add(ref.File);
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
			<div className={styles.container}>
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
							if (!this.state.entrySpecsByName || (!this.state.showAllFiles && i >= this.state.fileCollapseThreshold)) return null;

							let entrySpec = this.state.entrySpecsByName[loc.Path];
							if (!this.state.shownFiles[i]) return this.renderFileHeader(entrySpec, loc.Count, i);

							let file = this.state.filesByName ? this.state.filesByName[loc.Path] : null;
							if (!file) {
								let numLines = (SNIPPET_REF_CONTEXT_LINES * 2 * loc.Count) + (loc.Count * 2); // heuristic
								return <div key={i}>{this.renderFileHeader(entrySpec, loc.Count, i)}<BlobContentPlaceholder key={i} numLines={numLines} /></div>;
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
										annotations={this.state.anns[path] || null}
										activeDef={this.state.activeDef}
										activeDefNoRev={this.state.activeDef ? urlToDef(def, "") : null}
										lineNumbers={true}
										displayRanges={this.state.ranges[path] || null}
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
