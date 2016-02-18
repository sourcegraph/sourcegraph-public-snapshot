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
import IssueForm from "sourcegraph/issue/IssueForm";
import context from "sourcegraph/context";
import "sourcegraph/code/CodeBackend";
import "sourcegraph/def/DefBackend";

function lineFromByte(file, byte) {
	if (!file || !byte) { return null; }
	let lines = file.Entry.SourceCode.Lines;
	for (let i = 0; i < lines.length; i++) {
		if (lines[i].StartByte <= byte && byte <= lines[i].EndByte) {
			return i + 1;
		}
	}
	return null;
}

class CodeFileContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			creatingIssue: false,
		};
		this._onClick = this._onClick.bind(this);
		this._onKeyDown = this._onKeyDown.bind(this);
		this._onLineButtonClick = this._onLineButtonClick.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		document.addEventListener("click", this._onClick);
		document.addEventListener("keydown", this._onKeyDown);

		// FIXME: jQuery is deprecated.
		window.jQuery(this.refs.toolbar).fixedsticky();
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

		// get filename from definition data
		let defData = props.def && DefStore.defs.get(props.def);
		state.tree = props.def ? (defData && defData.File.Path) : props.tree;
		state.selectedDef = props.def || props.selectedDef; // triggers WantDef for props.def

		// fetch file content
		state.file = state.tree && CodeStore.files.get(state.repo, state.rev, state.tree);

		state.startLine = props.def ? lineFromByte(state.file, defData && defData.ByteStartPosition) : (props.startLine || null);
		state.endLine = props.def ? lineFromByte(state.file, defData && defData.ByteEndPosition) : (props.endLine || null);

		state.defs = DefStore.defs;
		state.examples = DefStore.examples;
		state.highlightedDef = DefStore.highlightedDef;

		state.defOptionsURLs = DefStore.defOptionsURLs;
		state.defOptionsLeft = DefStore.defOptionsLeft;
		state.defOptionsTop = DefStore.defOptionsTop;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.tree && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.tree !== nextState.tree)) {
			Dispatcher.asyncDispatch(new CodeActions.WantFile(nextState.repo, nextState.rev, nextState.tree));
		}
		if (nextState.selectedDef && prevState.selectedDef !== nextState.selectedDef) {
			Dispatcher.asyncDispatch(new DefActions.WantDef(nextState.selectedDef));
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

	_onLineButtonClick(lineNumber, selected) {
		this.setState({creatingIssue: true}, () => {
			if (!selected) {
				Dispatcher.dispatch(new CodeActions.SelectLine(lineNumber));
			}
		});
	}

	render() {
		if (!this.state.tree) {
			return null;
		}

		let selectedDefData = this.state.selectedDef && this.state.defs.get(this.state.selectedDef);
		let highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);
		return (
			<div className="row">
				<CodeFileToolbar
					repo={this.state.repo}
					rev={this.state.rev}
					tree={this.state.tree}
					file={this.state.file} />
				{this.state.file &&
					<div className="code-view-react">
						<CodeListing
							lines={this.state.file.Entry.SourceCode.Lines}
							lineNumbers={true}
							startLine={this.state.startLine}
							endLine={this.state.endLine}
							selectedDef={this.state.selectedDef}
							highlightedDef={this.state.highlightedDef}
							onLineButtonClick={!context.isMothership ? this._onLineButtonClick : null}
							lineSelectionForm={
								(this.state.creatingIssue && this.state.startLine && this.state.endLine) ? (
								<IssueForm
									repo={this.state.repo}
									path={this.state.tree}
									commitID={this.state.file.EntrySpec.RepoRev.CommitID}
									startLine={this.state.startLine}
									endLine={this.state.endLine}
									onCancel={() => { this.setState({creatingIssue: false}); }}
									onSubmit={(url) => {
										this.setState({creatingIssue: false});
										window.location.href = url;
									}} />
							) : null} />
					</div>
				}

				{selectedDefData &&
					<DefPopup
						def={selectedDefData}
						examples={this.state.examples}
						highlightedDef={this.state.highlightedDef} />
				}

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
										Dispatcher.dispatch(new DefActions.SelectDef(url));
									}}>
										{data ? <span dangerouslySetInnerHTML={data.QualifiedName} /> : "..."}
									</li>
								);
							})}
						</ul>
					</div>
				}
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
	endLine: React.PropTypes.number,
	selectedDef: React.PropTypes.string,
	example: React.PropTypes.number,
};

export default CodeFileContainer;
