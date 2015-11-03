import React from "react";

import Container from "./Container";
import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import CodeListing from "./CodeListing";
import DefPopup from "./DefPopup";
import "./CodeBackend";
import "./DefBackend";

export default class CodeFileContainer extends Container {
	stores() {
		return [CodeStore, DefStore];
	}

	reconcileState(state, props) {
		Object.assign(state, props);

		state.file = CodeStore.files.get(state.repo, state.rev, state.tree);

		state.defs = DefStore.defs;
		state.defsGeneration = DefStore.defs.generation;
		state.examples = DefStore.examples;
		state.examplesGeneration = DefStore.examples.generation;
		state.highlightedDef = DefStore.highlightedDef;
		state.discussions = DefStore.discussions;
		state.discussionsGeneration = DefStore.discussions.generation;
	}

	requestData(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.tree !== nextState.tree) {
			Dispatcher.dispatch(new CodeActions.WantFile(nextState.repo, nextState.rev, nextState.tree));
		}
		if (nextState.selectedDef && prevState.selectedDef !== nextState.selectedDef) {
			Dispatcher.dispatch(new DefActions.WantDef(nextState.selectedDef));
			Dispatcher.dispatch(new DefActions.WantDiscussions(nextState.selectedDef));
		}
	}

	render() {
		if (!this.state.file) {
			return null;
		}
		let def = this.state.selectedDef && this.state.defs.get(this.state.selectedDef);
		return (
			<div>
				<div className="code-view-react">
					<CodeListing
						lines={this.state.file.Entry.SourceCode.Lines}
						lineNumbers={true}
						startLine={this.state.startLine}
						endLine={this.state.endLine}
						selectedDef={this.state.selectedDef}
						highlightedDef={this.state.highlightedDef} />
				</div>
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
	startLine: React.PropTypes.number,
	endLine: React.PropTypes.number,
	selectedDef: React.PropTypes.string,
	unitType: React.PropTypes.string,
	unit: React.PropTypes.string,
	def: React.PropTypes.string,
	example: React.PropTypes.number,
};
