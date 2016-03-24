import React from "react";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import hotLink from "sourcegraph/util/hotLink";
import * as router from "sourcegraph/util/router";

class ExamplesContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			currentPage: 1,
		};
	}

	stores() {
		return [DefStore, BlobStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.defs = DefStore.defs;
		state.refs = DefStore.refs.get(state.def);
		state.annotations = BlobStore.annotations;
		state.examples = DefStore.examples.get(state.def, state.currentPage);
		state.files = [];
		state.ranges = {};
		state.anns = {};

		let fileIndex = new Set();
		for (let ex of state.examples || []) {
			if (!fileIndex.has(ex.File)) {
				state.files.push(BlobStore.files.get(ex.Repo, ex.Rev, ex.File));
				state.ranges[ex.File] = [];
				fileIndex.add(ex.File);
			}
			state.ranges[ex.File].push([ex.Range.StartLine, ex.Range.EndLine]);
			let anns = state.annotations.get(ex.Repo, ex.Rev, "", ex.File);
			state.anns[ex.File] = anns ? anns.Annotations : null;
		}

		state.highlightedDef = DefStore.highlightedDef || null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.def && prevState.def !== nextState.def) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.def));
			Dispatcher.asyncDispatch(new DefActions.WantRefs(nextState.def));
			Dispatcher.asyncDispatch(new DefActions.WantExamples(nextState.def, 1));
		}

		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.highlightedDef));
		}

		if (nextState.examples && (prevState.examples && prevState.examples.length) !== nextState.examples) {
			let wantedFiles = new Set();
			for (let ex of nextState.examples) {
				if (wantedFiles.has(ex.File)) continue; // Prevent many requests for the same file.
				// TODO Only fetch a portion of the file/annotations at a time for perf.
				Dispatcher.asyncDispatch(new BlobActions.WantFile(ex.Repo, ex.Rev, ex.File));
				Dispatcher.asyncDispatch(new BlobActions.WantAnnotations(ex.Repo, ex.Rev, "", ex.File));
				wantedFiles.add(ex.File);
			}
		}
	}

	render() {
		let defData = this.state.def && this.state.defs.get(this.state.def);
		let highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);

		return (
			<div>
				<header>Examples for {defData && <a href={defData.URL} onClick={hotLink} dangerouslySetInnerHTML={defData.QualifiedName}/>}</header>
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
						</div>
					</div>
				</div>

				{highlightedDefData && <DefTooltip currentRepo={this.state.repo} def={highlightedDefData} />}
			</div>
		);
	}
}

export default ExamplesContainer;
