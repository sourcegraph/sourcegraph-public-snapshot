import React from "react";

import Container from "./Container";
import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import CodeListing from "./CodeListing";
import DefPopup from "./DefPopup";
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
	}

	requestData(prevState, nextState) {
		if (nextState.tree && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.tree !== nextState.tree)) {
			Dispatcher.dispatch(new CodeActions.WantFile(nextState.repo, nextState.rev, nextState.tree));
		}
		if (nextState.selectedDef && prevState.selectedDef !== nextState.selectedDef) {
			Dispatcher.dispatch(new DefActions.WantDef(nextState.selectedDef));
			Dispatcher.dispatch(new DefActions.WantDiscussions(nextState.selectedDef));
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

		let def = this.state.selectedDef && this.state.defs.get(this.state.selectedDef);
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
				{def &&
					<DefPopup
						def={def}
						examples={this.state.examples}
						highlightedDef={this.state.highlightedDef}
						discussions={this.state.discussions.get(this.state.selectedDef)} />
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
