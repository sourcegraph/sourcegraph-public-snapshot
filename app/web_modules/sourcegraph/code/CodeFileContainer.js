import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as CodeActions from "sourcegraph/code/CodeActions";
import * as DefActions from "sourcegraph/def/DefActions";
import CodeStore from "sourcegraph/code/CodeStore";
import DefStore from "sourcegraph/def/DefStore";
import CodeListing from "sourcegraph/code/CodeListing";
import CodeFileToolbar from "sourcegraph/code/CodeFileToolbar";
import DefPopup from "sourcegraph/def/DefPopup";
import DefTooltip from "sourcegraph/def/DefTooltip";
import FileMargin from "sourcegraph/code/FileMargin";
import "sourcegraph/code/CodeBackend";
import "sourcegraph/def/DefBackend";
import lineFromByte from "sourcegraph/code/lineFromByte";
import {GoTo} from "sourcegraph/util/hotLink";

class CodeFileContainer extends Container {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
		this._onKeyDown = this._onKeyDown.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		document.addEventListener("click", this._onClick);
		document.addEventListener("keydown", this._onKeyDown);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		document.removeEventListener("click", this._onClick);
		document.removeEventListener("keydown", this._onKeyDown);
	}

	stores() {
		return [CodeStore, DefStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.activeDef = props.activeDef || null;
		let defData = props.activeDef && DefStore.defs.get(state.activeDef);
		state.tree = props.activeDef && defData ? defData.File.Path : props.tree;

		// fetch file content
		state.file = state.tree && CodeStore.files.get(state.repo, state.rev, state.tree);
		state.anns = state.tree && CodeStore.annotations.get(state.repo, state.rev, state.tree, 0, 0);
		state.annotations = CodeStore.annotations;

		state.startLine = (state.activeDef && state.file && defData) ? lineFromByte(state.file.Entry.ContentsString, defData && defData.ByteStartPosition) : (props.startLine || null);

		state.defs = DefStore.defs;
		state.examples = DefStore.examples;
		state.highlightedDef = DefStore.highlightedDef || null;

		state.defOptionsURLs = DefStore.defOptionsURLs;
		state.defOptionsLeft = DefStore.defOptionsLeft;
		state.defOptionsTop = DefStore.defOptionsTop;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.tree && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.tree !== nextState.tree)) {
			Dispatcher.asyncDispatch(new CodeActions.WantFile(nextState.repo, nextState.rev, nextState.tree));
			Dispatcher.asyncDispatch(new CodeActions.WantAnnotations(nextState.repo, nextState.rev, nextState.tree));
		}
		if (nextState.activeDef && prevState.activeDef !== nextState.activeDef) {
			let defData = nextState.activeDef && DefStore.defs.get(nextState.activeDef);
			if (defData && (!defData.File.Path || (defData.Data && defData.Data.Kind === "package"))) {
				// The def's File field refers to a directory (e.g., in the
				// case of a Go package). We can't show a dir in this view,
				// so just redirect to the dir listing.
				//
				// TODO(sqs): Improve handling of this case.
				window.location.href = defData.URL;
				return;
			}
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.activeDef));
		}
		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.highlightedDef));
		}
		if (nextState.defOptionsURLs && prevState.defOptionsURLs !== nextState.defOptionsURLs) {
			nextState.defOptionsURLs.forEach((url) => {
				Dispatcher.asyncDispatch(new DefActions.WantDef(url));
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
		if (!this.state.tree) {
			return null;
		}

		let defData = this.state.activeDef && this.state.defs.get(this.state.activeDef);
		let highlightedDefData = this.state.highlightedDef && this.state.highlightedDef !== this.state.activeDef && this.state.defs.get(this.state.highlightedDef);

		return (
			<div className="file-container">
				<div className="content-view">
					<div className="content file-content card">
						<CodeFileToolbar
							repo={this.state.repo}
							rev={this.state.rev}
							tree={this.state.tree}
							file={this.state.file} />
						{this.state.file &&
						<CodeListing
							ref={(e) => this.setState({_codeListing: e})}
							contents={this.state.file.Entry.ContentsString}
							annotations={this.state.anns ? this.state.anns.Annotations : null}
							lineNumbers={true}
							startLine={this.state.startLine}
							startCol={this.state.startCol}
							endLine={this.state.endLine}
							endCol={this.state.endCol}
							highlightedDef={this.state.highlightedDef}
							activeDef={this.state.activeDef} />}
					</div>
					<FileMargin getOffsetTopForByte={this.state._codeListing ? this.state._codeListing.getOffsetTopForByte.bind(this.state._codeListing) : null}>
						{defData &&
						<DefPopup
							def={defData}
							byte={defData.ByteStartPosition}
							examples={this.state.examples}
							annotations={this.state.annotations}
							activeDef={this.state.activeDef}
							highlightedDef={this.state.highlightedDef} />}
					</FileMargin>
				</div>

				{highlightedDefData && highlightedDefData.Found && !this.state.defOptionsURLs && <DefTooltip def={highlightedDefData} />}

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
								<li key={i} onClick={() => {
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

CodeFileContainer.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	tree: React.PropTypes.string,
	def: React.PropTypes.string,
	startLine: React.PropTypes.number,
	startCol: React.PropTypes.number,
	endLine: React.PropTypes.number,
	endCol: React.PropTypes.number,
	example: React.PropTypes.number,
};

export default CodeFileContainer;
