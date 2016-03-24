import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import BuildStore from "sourcegraph/build/BuildStore";
import DefStore from "sourcegraph/def/DefStore";
import Blob from "sourcegraph/blob/Blob";
import BlobToolbar from "sourcegraph/blob/BlobToolbar";
import DefPopup from "sourcegraph/def/DefPopup";
import DefTooltip from "sourcegraph/def/DefTooltip";
import FileMargin from "sourcegraph/blob/FileMargin";
import "sourcegraph/blob/BlobBackend";
import "sourcegraph/def/DefBackend";
import "sourcegraph/build/BuildBackend";
import lineFromByte from "sourcegraph/blob/lineFromByte";
import {GoTo} from "sourcegraph/util/hotLink";

class BlobContainer extends Container {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
		this._onKeyDown = this._onKeyDown.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		if (typeof document !== "undefined") {
			document.addEventListener("click", this._onClick);
			document.addEventListener("keydown", this._onKeyDown);
		}
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (typeof document !== "undefined") {
			document.removeEventListener("click", this._onClick);
			document.removeEventListener("keydown", this._onKeyDown);
		}
	}

	stores() {
		return [BuildStore, BlobStore, BuildStore, DefStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.activeDef = props.activeDef || null;
		let activeDefData = props.activeDef && DefStore.defs.get(state.activeDef);

		state.path = props.activeDef && activeDefData ? activeDefData.File : props.path;

		// fetch file content
		state.file = state.path ? BlobStore.files.get(state.repo, state.rev, state.path) : null;
		state.anns = state.path ? BlobStore.annotations.get(state.repo, state.rev, "", state.path, 0, 0) : null;
		state.annotations = BlobStore.annotations;

		if (state.activeDef && state.file && activeDefData) {
			state.startLine = lineFromByte(state.file.ContentsString, activeDefData.DefStart);
			state.endLine = lineFromByte(state.file.ContentsString, activeDefData.DefEnd);
		}

		state.defs = DefStore.defs;
		state.refs = DefStore.refs.get(state.activeDef);
		state.highlightedDef = DefStore.highlightedDef || null;

		state.defOptionsURLs = DefStore.defOptionsURLs;
		state.defOptionsLeft = DefStore.defOptionsLeft;
		state.defOptionsTop = DefStore.defOptionsTop;

		state.builds = BuildStore.builds;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.path && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path)) {
			Dispatcher.asyncDispatch(new BlobActions.WantFile(nextState.repo, nextState.rev, nextState.path));
			Dispatcher.asyncDispatch(new BlobActions.WantAnnotations(nextState.repo, nextState.rev, "", nextState.path));
		}
		if (nextState.activeDef && prevState.activeDef !== nextState.activeDef) {
			let activeDefData = nextState.activeDef && DefStore.defs.get(nextState.activeDef);
			if (activeDefData && (!activeDefData.File || activeDefData.Kind === "package")) {
				// The def's File field refers to a directory (e.g., in the
				// case of a Go package). We can't show a dir in this view,
				// so just redirect to the dir listing.
				//
				// TODO(sqs): Improve handling of this case.
				if (typeof window !== "undefined") window.location.href = activeDefData.URL;
				return;
			}
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.activeDef));
			Dispatcher.asyncDispatch(new DefActions.WantRefs(nextState.activeDef));
		}
		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.highlightedDef));
		}
		if (nextState.defOptionsURLs && prevState.defOptionsURLs !== nextState.defOptionsURLs) {
			nextState.defOptionsURLs.forEach((url) => {
				Dispatcher.dispatch(new DefActions.WantDef(url));
			});
		}
	}

	_onClick() {
		if (this.state.defOptionsURLs) {
			Dispatcher.dispatch(new DefActions.SelectMultipleDefs(null, 0, 0));
		}
	}

	_onKeyDown(event) {
		if (event.keyCode === 27) {
			Dispatcher.dispatch(new DefActions.SelectDef(null));
		}
	}

	render() {
		if (!this.state.path) {
			return null;
		}

		let activeDefData = this.state.activeDef && this.state.defs.get(this.state.activeDef);
		let highlightedDefData = this.state.highlightedDef && this.state.highlightedDef !== this.state.activeDef && this.state.defs.get(this.state.highlightedDef);

		return (
			<div className="file-container">
				<div className="content-view">
					<div className="content file-content card">
						<BlobToolbar
							builds={this.state.builds}
							repo={this.state.repo}
							rev={this.state.rev}
							path={this.state.path} />
						{this.state.file &&
						<Blob
							ref={(e) => this._blob = e}
							contents={this.state.file.ContentsString}
							annotations={this.state.anns}
							lineNumbers={true}
							highlightSelectedLines={true}
							startLine={this.state.startLine}
							startCol={this.state.startCol}
							endLine={this.state.endLine}
							endCol={this.state.endCol}
							scrollToStartLine={true}
							highlightedDef={this.state.highlightedDef}
							activeDef={this.state.activeDef}
							dispatchSelections={true} />}
					</div>
					<FileMargin getOffsetTopForByte={this._blob ? this._blob.getOffsetTopForByte.bind(this._blob) : null}>
						{activeDefData && !activeDefData.Error &&
						<DefPopup
							def={activeDefData}
							byte={activeDefData.DefStart}
							refs={this.state.refs}
							path={this.state.path}
							annotations={this.state.annotations}
							activeDef={this.state.activeDef}
							highlightedDef={this.state.highlightedDef} />}
					</FileMargin>
				</div>

				{highlightedDefData && !this.state.defOptionsURLs && <DefTooltip currentRepo={this.state.repo} def={highlightedDefData} />}

				{this.state.defOptionsURLs &&
				<div className="context-menu"
					style={{
						left: this.state.defOptionsLeft,
						top: this.state.defOptionsTop,
					}}>
					<ul>
						{this.state.defOptionsURLs.map((url, i) => {
							let data = this.state.defs.get(url);
							return (
								<li key={i} onClick={(ev) => {
									Dispatcher.dispatch(new GoTo(url));
								}}>
									{data ? <span dangerouslySetInnerHTML={data.QualifiedName} /> : "..."}
								</li>
							);
						})}
					</ul>
				</div>}
			</div>
		);
	}
}

BlobContainer.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	path: React.PropTypes.string,
	def: React.PropTypes.string,
	startLine: React.PropTypes.number,
	startCol: React.PropTypes.number,
	endLine: React.PropTypes.number,
	endCol: React.PropTypes.number,
};

export default BlobContainer;
