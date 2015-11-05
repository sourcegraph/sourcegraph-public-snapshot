import React from "react";

import Container from "./Container";
import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import CodeListing from "./CodeListing";
import DefPopup from "./DefPopup";
import DefTooltip from "./DefTooltip";
import RepoBuildIndicator from "../components/RepoBuildIndicator"; // FIXME
import RepoRevSwitcher from "../components/RepoRevSwitcher"; // FIXME
import "./CodeBackend";
import "./DefBackend";

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

export default class CodeFileContainer extends Container {
	constructor(props) {
		super(props);
		this._hideOptionsMenu = this._hideOptionsMenu.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		document.addEventListener("click", this._hideOptionsMenu);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		document.removeEventListener("click", this._hideOptionsMenu);
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
		state.defsGeneration = DefStore.defs.generation;
		state.examples = DefStore.examples;
		state.examplesGeneration = DefStore.examples.generation;
		state.highlightedDef = DefStore.highlightedDef;
		state.discussions = DefStore.discussions;
		state.discussionsGeneration = DefStore.discussions.generation;

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
			Dispatcher.asyncDispatch(new DefActions.WantDiscussions(nextState.selectedDef));
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

	_hideOptionsMenu() {
		if (this.state.defOptionsURLs) {
			Dispatcher.dispatch(new DefActions.SelectMultipleDefs(null, 0, 0));
		}
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
							highlightedDef={this.state.highlightedDef} />
					</div>
				}

				{selectedDefData &&
					<DefPopup
						def={selectedDefData}
						examples={this.state.examples}
						highlightedDef={this.state.highlightedDef}
						discussions={this.state.discussions.get(this.state.selectedDef)} />
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
