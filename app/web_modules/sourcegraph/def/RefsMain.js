// @flow weak

import React from "react";

import Blob from "sourcegraph/blob/Blob";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import {Link} from "react-router";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/blob/BlobBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {urlToDef} from "sourcegraph/def/routes";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/Refs.css";

const FILES_PER_PAGE = 5;

class RefsMain extends Container {
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
		state.repo = props.repo;
		state.rev = props.rev;
		state.def = props.def;
		state.defObj = props.defObj;
		state.path = props.location.query.file || null;
		state.refs = DefStore.refs.get(state.repo, state.rev, state.def, state.path);
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
			for (let ref of state.refs.Refs || []) {
				if (!ref) continue;
				let refRev = ref.Repo === state.repo ? state.rev : ref.CommitID;
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
				anns[ref.File] = BlobStore.annotations.get(ref.Repo, refRev, ref.CommitID, ref.File);
			}
			state.files = files;
			state.entrySpecs = entrySpecs;
			state.ranges = ranges;
			state.anns = anns;
		}

		state.highlightedDef = DefStore.highlightedDef || null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.def && prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.path !== nextState.path) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.path));
		}


		if (nextState.refs && (nextState.refs !== prevState.refs || nextState.page !== prevState.page)) {
			let wantedFiles = new Set();
			for (let ref of nextState.refs.Refs) {
				if (wantedFiles.size === (nextState.page * FILES_PER_PAGE)) break;
				if (wantedFiles.has(ref.File)) continue; // Prevent many requests for the same file.
				// TODO Only fetch a portion of the file/annotations at a time for perf.
				let refRev = ref.Repo === nextState.repo ? nextState.rev : ref.CommitID;
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(ref.Repo, refRev, ref.File));
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(ref.Repo, refRev, ref.CommitID, ref.File));
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
		let maxFilesShown = this.state.page * FILES_PER_PAGE;

		return (
			<div styleName="refs-container">
				<h1>Refs to {this.state.defObj && <Link to={urlToDef(this.state.defObj)} dangerouslySetInnerHTML={this.state.defObj.QualifiedName}/>} {this.state.path ? `in ${this.state.path}` : `in ${this.state.repo}`}</h1>
				<hr/>
				{this.state.files && this.state.files.map((file, i) => {
					if (!file) return null;
					let entrySpec = this.state.entrySpecs[i];
					let path = entrySpec.Path;
					let repoRev = entrySpec.RepoRev;
					return (
						<div key={path}>
							<h3>
								<i className="fa fa-file"/>
								<Link to={urlToBlob(repoRev.URI, repoRev.Rev, path)}>{path}</Link>
							</h3>
							<Blob
								repo={repoRev.URI}
								rev={repoRev.Rev}
								path={path}
								contents={file.ContentsString}
								annotations={this.state.anns[path] || null}
								activeDef={this.state.def}
								lineNumbers={true}
								displayRanges={this.state.ranges[path] || null}
								highlightedDef={this.state.highlightedDef} />
						</div>
					);
				})}
				{this.state.files && this.state.files.length > maxFilesShown &&
					<div styleName="refs-footer">
						<span styleName="search-hotkey" data-hint={`Refs from ${maxFilesShown} out of ${this.state.files.length} files currently shown`}><button onClick={this._nextPage}>View more</button></span>
					</div>
				}

				{this.state.highlightedDefObj && <DefTooltip currentRepo={this.state.repo} def={this.state.highlightedDefObj} />}
			</div>
		);
	}
}

export default CSSModules(RefsMain, styles);
