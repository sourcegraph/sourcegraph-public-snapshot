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
import {urlToDef, urlToDef2} from "sourcegraph/def/routes";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {urlToBlob} from "sourcegraph/blob/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/Refs.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import RefLocationsList from "sourcegraph/def/RefLocationsList";
import {FileIcon} from "sourcegraph/components/Icons";
import Header from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/app/status";

const FILES_PER_PAGE = 5;

class RefsMain extends Container {
	static contextTypes = {
		status: React.PropTypes.object,
		router: React.PropTypes.object.isRequired,
	};

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
		if (super.componentDidMount) super.componentDidMount();

		// Fix a bug where navigating from a blob page here does not cause the
		// browser to scroll to the top of this page.
		if (typeof window !== "undefined") window.scrollTo(0, 0);
	}

	componentWillUnmount() {
		if (super.componentWillUnmount) super.componentWillUnmount();
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.activeDef = state.def ? urlToDef2(state.repo, state.rev, state.def) : state.def;

		state.refLocations = state.def ? DefStore.refLocations.get(state.repo, state.rev, state.def) : null;

		state.refRepo = props.location && props.location.query.repo ? props.location.query.repo : null;
		state.refFile = props.location && props.location.query.file ? props.location.query.file : null;
		if (!state.refRepo && state.refLocations && !state.refLocations.Error && state.refLocations.length > 0) {
			state.refRepo = state.refLocations[0].Repo;
		}

		state.refs = props.refs || DefStore.refs.get(state.repo, state.rev, state.def, state.refRepo, state.refFile);
		state.files = null;
		state.entrySpecs = null;
		state.ranges = null;
		state.anns = null;
		state.highlightedDef = props.highlightedDef || null;
		state.highlightedDefObj = props.highlightedDefObj || null;

		if (state.refs && !state.refs.Error) {
			let files = [];
			let entrySpecs = [];
			let ranges = {};
			let anns = {};
			let fileIndex = new Map();
			for (let ref of state.refs || []) {
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
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations(nextState.repo, nextState.rev, nextState.def));
		}

		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def || prevState.refRepo !== nextState.refRepo || prevState.refFile !== nextState.refFile) {
			if (nextState.refRepo) {
				Dispatcher.Backends.dispatch(new DefActions.WantRefs(nextState.repo, nextState.rev, nextState.def, nextState.refRepo, nextState.refFile));
			}
		}

		if (nextState.refs && prevState.refs !== nextState.refs) {
			this.context.status.error(nextState.refs.Error);
		}

		if (nextState.refLocations && prevState.refLocations !== nextState.refLocations) {
			this.context.status.error(nextState.refLocations.Error);
		}

		if (nextState.refs && !nextState.refs.Error && (nextState.refs !== prevState.refs || nextState.page !== prevState.page)) {
			let wantedFiles = new Set();
			for (let ref of nextState.refs) {
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
		let def = this.state.defObj;
		let refLocs = this.state.refLocations;

		if (refLocs && refLocs.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocs.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		let maxFilesShown = this.state.page * FILES_PER_PAGE;

		return (
			<div styleName="refs-container">
				<h1>Refs to {this.state.defObj && <Link to={`${urlToDef(this.state.defObj)}/-/info`}><code>{qualifiedNameAndType(this.state.defObj)}</code></Link>}</h1>
				<div styleName="inner">
					<div styleName="ref-locations">
						{def && !def.Error && refLocs && !refLocs.Error && refLocs.length > 0 && <RefLocationsList def={def} refLocations={refLocs} repo={this.state.refRepo} path={this.state.refFile} />}
					</div>
					<div styleName="refs">
						{this.state.files && this.state.files.map((file, i) => {
							if (!file) return null;
							let entrySpec = this.state.entrySpecs[i];
							let path = entrySpec.Path;
							let repoRev = entrySpec.RepoRev;
							return (
								<div key={path}>
									<h3 styleName="file-name">
										<Link to={urlToBlob(repoRev.URI, repoRev.Rev, path)}>
											<FileIcon /> {path}
										</Link>
									</h3>
									<Blob
										repo={repoRev.URI}
										rev={repoRev.Rev}
										path={path}
										contents={file.ContentsString}
										annotations={this.state.anns[path] || null}
										activeDef={this.state.activeDef}
										activeDefNoRev={this.state.activeDef ? urlToDef(def, "") : null}
										lineNumbers={true}
										displayRanges={this.state.ranges[path] || null}
										highlightedDef={this.state.highlightedDef}
										highlightedDefObj={this.state.highlightedDefObj} />
								</div>
							);
						})}
						{this.state.files && this.state.files.length > maxFilesShown &&
							<div styleName="refs-footer">
								<span styleName="search-hotkey" data-hint={`Refs from ${maxFilesShown} out of ${this.state.files.length} files currently shown`}><button onClick={this._nextPage}>View more</button></span>
							</div>
						}
					</div>
				</div>
				{this.state.highlightedDefObj && !this.state.highlightedDefObj.Error && <DefTooltip currentRepo={this.state.repo} def={this.state.highlightedDefObj} />}
			</div>
		);
	}
}

export default CSSModules(RefsMain, styles);
