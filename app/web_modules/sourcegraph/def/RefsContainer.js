import React from "react";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import animatedScrollTo from "sourcegraph/util/animatedScrollTo";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import hotLink from "sourcegraph/util/hotLink";
import * as router from "sourcegraph/util/router";

import RefStyles from "sourcegraph/def/styles/Refs.css";

const FILES_PER_PAGE = 5;

function annsKeyFor(repo, file) {
	return `${repo || ""}:${file || ""}`;
}

class RefsContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			// Pagination limits the amount of files that are initiallly loaded to
			// prevent a flood of large requests.
			// TODO: This is only set when the component is created, which means if
			// you navigate to refs for another def, the page will not be reset.
			page: 1,
		};
		this._nextPage = this._nextPage.bind(this);
	}

	stores() {
		return [DefStore, BlobStore];
	}

	componentDidMount() {
		super.componentDidMount();
		animatedScrollTo(document.body, 0, 100);
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.defs = DefStore.defs;
		state.refs = DefStore.refs.get(state.def);
		state.refLocations = DefStore.refLocations.get(state.def);
		state.annotations = BlobStore.annotations;
		state.refRepo = props.refRepo || "";
		state.refFile = props.refFile || "";
		// refs holds all references to 'def' from repo 'refRepo' and file 'refFile'.
		// If 'refFile' is empty, refs contains references from all files in 'refRepo'.
		state.refs = DefStore.refs.get(state.def, state.refRepo, state.refFile);
		state.files = null;
		state.entrySpecs = null;
		state.ranges = null;
		state.anns = null;

		if (state.refs) {
			let files = [];
			let entrySpecs = [];
			let ranges = {};
			let anns = {};
			let fileIndex = new Map();
			for (let ref of state.refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.rev : "";
				if (!fileIndex.has(ref.File)) {
					let file = BlobStore.files.get(ref.Repo, refRev, ref.File);
					files.push(file);
					entrySpecs.push({RepoRev: {URI: ref.Repo, Rev: refRev}, Path: ref.File});
					ranges[ref.File] = [];
					fileIndex.set(ref.File, file);
				}
				let file = fileIndex.get(ref.File);
				// Determine the line range that should be displayed for each ref.
				if (file) {
					const context = 4; // Number of additional lines to show above/below a ref
					let contents = file.ContentsString;
					ranges[ref.File].push([
						Math.max(lineFromByte(contents, ref.Start) - context, 0),
						lineFromByte(contents, ref.End) + context,
					]);
				}
				anns[annsKeyFor(ref.Repo, ref.File)] = state.annotations.get(ref.Repo, refRev, "", ref.File);
			}
			state.files = files;
			state.entrySpecs = entrySpecs;
			state.ranges = ranges;
			state.anns = anns;
		}

		state.highlightedDef = DefStore.highlightedDef || null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.def && prevState.def !== nextState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.def));
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations(nextState.def));
		}

		if ((nextState.refFile && prevState.refFile !== nextState.refFile) ||
			(nextState.refRepo && prevState.refRepo !== nextState.refRepo)) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.def, nextState.refRepo, nextState.refFile));
		}

		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.highlightedDef));
		}

		if (nextState.refs && (nextState.refs !== prevState.refs || nextState.page !== prevState.page)) {
			let wantedFiles = new Set();
			for (let ref of nextState.refs || []) {
				if (wantedFiles.size === (nextState.page * FILES_PER_PAGE)) break;
				if (wantedFiles.has(ref.File)) continue; // Prevent many requests for the same file.
				// TODO Only fetch a portion of the file/annotations at a time for perf.
				let refRev = ref.Repo === nextState.repo ? nextState.rev : "";
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(ref.Repo, refRev, "", ref.File));
				wantedFiles.add(ref.File);
			}
		}
	}

	_nextPage() {
		this.setState({
			page: this.state.page + 1,
		});
	}

	render() {
		let defData = this.state.def && this.state.defs.get(this.state.def);
		let highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);
		let maxFilesShown = this.state.page * FILES_PER_PAGE;

		return (
			<div className={RefStyles.refs_container}>
				<header>Refs of {defData && <a href={defData.URL} onClick={hotLink} dangerouslySetInnerHTML={defData.QualifiedName}/>} {this.state.refFile && `in ${this.state.refFile} `}in {this.state.refRepo}</header>
				<hr/>
				<div className="file-container">
					<div className="content-view">
						<div className="content file-content">
							{!this.state.files && <i className="fa fa-circle-o-notch fa-spin"></i>}
							{this.state.files && this.state.files.map((file, i) => {
								if (!file) return null;
								let entrySpec = this.state.entrySpecs[i];
								let path = entrySpec.Path;
								let repoRev = entrySpec.RepoRev;
								return (
									<div className="card file-list-item" key={path}>
										<div className="code-file-toolbar" ref="toolbar">
											<div className="file-breadcrumb">
												<i className="fa fa-file"/>
												<a href={router.tree(repoRev.URI, repoRev.Rev, path)}>{path}</a>
											</div>
										</div>
										<Blob
											repo={repoRev.URI}
											rev={repoRev.Rev}
											path={path}
											contents={file.ContentsString}
											annotations={this.state.anns[annsKeyFor(repoRev.URI, path)] || null}
											activeDef={this.state.def}
											lineNumbers={true}
											displayRanges={this.state.ranges[path] || null}
											highlightedDef={this.state.highlightedDef} />
									</div>
								);
							})}
							{this.state.files && this.state.files.length > maxFilesShown &&
								<div className="refs-footer">
									<span className={RefStyles.search_hotkey} data-hint={`Refs from ${maxFilesShown} out of ${this.state.files.length} files currently shown`}><div className="btn btn-default" onClick={this._nextPage}>View more</div></span>
								</div>
							}
						</div>
					</div>
				</div>

				{highlightedDefData && <DefTooltip currentRepo={this.state.repo} def={highlightedDefData} />}
			</div>
		);
	}
}

export default RefsContainer;
