import React from "react";
import {Container} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import CodeStore from "./CodeStore";
import CodeListing from "./CodeListing";
import "./CodeBackend";

class CodeFileContainer extends React.Component {
	componentWillMount() {
		Dispatcher.dispatch(new CodeActions.WantFile(this.props.repo, this.props.rev, this.props.tree));
	}

	static getStores() {
		return [CodeStore];
	}

	static calculateState(prevState) {
		return {
			files: CodeStore.files,
			highlightedDef: CodeStore.highlightedDef,
		};
	}

	render() {
		let file = this.state.files.get(this.props.repo, this.props.rev, this.props.tree);
		if (file === undefined) {
			return null;
		}
		return (
			<CodeListing
				lines={file.Entry.SourceCode.Lines}
				selectedDef={this.props.selectedDef}
				highlightedDef={this.state.highlightedDef} />
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

export default Container.create(CodeFileContainer, {pure: false});
