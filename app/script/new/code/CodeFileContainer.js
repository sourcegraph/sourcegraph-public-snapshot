import React from "react";

import Container from "../Container";
import Dispatcher from "../Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "../def/DefActions";
import CodeStore from "./CodeStore";
import DefStore from "../def/DefStore";
import CodeListing from "./CodeListing";
import DefPopup from "../def/DefPopup";
import DefTooltip from "../def/DefTooltip";
import IssueForm from "../issue/IssueForm";
import RepoBuildIndicator from "../../components/RepoBuildIndicator"; // FIXME
import RepoRevSwitcher from "../../components/RepoRevSwitcher"; // FIXME
import "./CodeBackend";
import "../def/DefBackend";

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

		// TODO replace with proper shared component
		let basePath = `/${this.state.repo}@${this.state.rev}/.tree`;
		let repoSegs = this.state.repo.split("/");
		let breadcrumb = [<a key="base" href={basePath}>{repoSegs[repoSegs.length-1]}</a>];
		this.state.tree.split("/").forEach((seg, i) => {
			basePath += `/${seg}`;
			breadcrumb.push(<span key={i}> / <a href={basePath}>{seg}</a></span>);
		});


		let embedLink = `/${this.state.repo}@${this.state.rev}/.tree/${this.state.tree}/.share`;
		if (this.state.startLine && this.state.endLine) {
			embedLink += `?StartLine=${this.state.startLine}&EndLine=${this.state.endLine}`;
		}

		let selectedDefData = this.state.selectedDef && this.state.defs.get(this.state.selectedDef);
		let highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);
		return (
			<div>
				<div className="code-file-toolbar">
					<div className="file">
						<i className={this.state.file ? "fa fa-file" : "fa fa-spinner fa-spin"} />{breadcrumb}

						{this.state.file &&
							<RepoBuildIndicator
								RepoURI={this.state.repo}
								Rev={this.state.file.EntrySpec.RepoRev.CommitID}
								btnSize="btn-xs"
								tooltipPosition="bottom" />
						}
					</div>

					<div className="actions">
						<RepoRevSwitcher repoSpec={this.state.repo}
							rev={this.state.rev}
							path={this.state.tree}
							alignRight={true} />

						<a className="share top-action btn btn-default btn-xs"
							href={embedLink}
							data-tooltip={true} title="Select text to specify a line range">Embed</a>
					</div>
				</div>

				{this.state.file &&
					<div className="code-view-react">
						<CodeListing
							lines={this.state.file.Entry.SourceCode.Lines}
							lineNumbers={true}
							startLine={this.state.startLine}
							endLine={this.state.endLine}
							selectedDef={this.state.selectedDef}
							highlightedDef={this.state.highlightedDef}
							onLineButtonClick={this._onLineButtonClick}
							lineSelectionForm={(this.state.creatingIssue && this.state.startLine && this.state.endLine) ? (
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
