import React from "react";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import hotLink from "sourcegraph/util/hotLink";
import * as router from "sourcegraph/util/router";

import RefStyles from "sourcegraph/def/styles/Refs.css";

const FILES_PER_PAGE = 5;

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

	reconcileState(state, props) {
		Object.assign(state, props);

		state.defs = DefStore.defs;
		state.refs = DefStore.refs.get(state.def);
		state.annotations = BlobStore.annotations;
		state.path = props.path || "";
		state.refs = DefStore.refs.get(state.def, state.path);
		state.files = null;
		state.ranges = null;
		state.anns = null;

		if (state.refs) {
			let files = [];
			let ranges = {};
			let anns = {};
			let fileIndex = new Map();
			for (let ref of state.refs.Refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.rev : ref.CommitID;
				if (!fileIndex.has(ref.File)) {
					let file = BlobStore.files.get(ref.Repo, refRev, ref.File);
					files.push(file);
					ranges[ref.File] = [];
					fileIndex.set(ref.File, file);
				}
				let file = fileIndex.get(ref.File);
				// Determine the line range that should be displayed for each ref.
				if (file) {
					const context = 4; // Number of additional lines to show above/below a ref
					let contents = file.Entry.ContentsString;
					ranges[ref.File].push([
						Math.max(lineFromByte(contents, ref.Start) - context, 0),
						lineFromByte(contents, ref.End) + context,
					]);
				}
				let fileAnns = state.annotations.get(ref.Repo, refRev, ref.CommitID, ref.File);
				anns[ref.File] = fileAnns ? fileAnns.Annotations : null;
			}
			state.files = files;
			state.ranges = ranges;
			state.anns = anns;
		}

		state.highlightedDef = DefStore.highlightedDef || null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.def && prevState.def !== nextState.def) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.def));
			Dispatcher.asyncDispatch(new DefActions.WantRefs(nextState.def, nextState.path));
		}

		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.highlightedDef));
		}

		if (nextState.refs && (nextState.refs !== prevState.refs || nextState.page !== prevState.page)) {
			let wantedFiles = new Set();
			for (let ref of nextState.refs.Refs) {
				if (wantedFiles.size === (nextState.page * FILES_PER_PAGE)) break;
				if (wantedFiles.has(ref.File)) continue; // Prevent many requests for the same file.
				// TODO Only fetch a portion of the file/annotations at a time for perf.
				let refRev = ref.Repo === nextState.repo ? nextState.rev : ref.CommitID;
				Dispatcher.asyncDispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
				Dispatcher.asyncDispatch(new BlobActions.WantAnnotations(ref.Repo, refRev, ref.CommitID, ref.File));
				wantedFiles.add(ref.File);
			}
		}
	}

	_nextPage() {
		this.setState({
			page: this.state.page + 1,
		});
		console.log("FILS", this.state.files);
	}

	render() {
		let defData = this.state.def && this.state.defs.get(this.state.def);
		let highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);
		let maxFilesShown = this.state.page * FILES_PER_PAGE;

		return (
			<div>
				<header>Refs of {defData && <a href={defData.URL} onClick={hotLink} dangerouslySetInnerHTML={defData.QualifiedName}/>} {this.state.path ? `in ${this.state.path}` : `in ${this.state.repo}`}</header>
				<hr/>
				<div className="file-container">
					<div className="content-view">
						<div className="content file-content">
							{this.state.files && this.state.files.map((file, i) => {
								if (!file) return null;
								let path = file.EntrySpec.Path;
								let repoRev = file.EntrySpec.RepoRev;
								return (
									<div className="card file-list-item" key={path}>
										<div className="code-file-toolbar" ref="toolbar">
											<div className="file-breadcrumb">
												<i className="fa fa-file"/>
												<a href={router.tree(repoRev.URI, repoRev.Rev, path)}>{path}</a>
											</div>
										</div>
										<Blob
											contents={file.Entry.ContentsString}
											annotations={this.state.anns[path] || null}
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
