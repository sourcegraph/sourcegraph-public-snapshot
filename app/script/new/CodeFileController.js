import React from "react";
import {Container} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import CodeStore from "./CodeStore";
import CodeListing from "./CodeListing";
import "./CodeBackend";

class CodeFileController extends React.Component {
	componentWillMount() {
		Dispatcher.dispatch(new CodeActions.WantFile(this.props.repo, this.props.rev, this.props.tree));
	}

	static getStores() {
		return [CodeStore];
	}

	static calculateState(prevState) {
		return {
			files: CodeStore.files,
		};
	}

	render() {
		let file = this.state.files.get(this.props.repo, this.props.rev, this.props.tree);
		if (file === undefined) {
			return null;
		}
		return (
			<CodeListing lines={file.Entry.SourceCode.Lines} />
		);
	}
}

CodeFileController.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	tree: React.PropTypes.string,
	startline: React.PropTypes.number,
	endline: React.PropTypes.number,
	token: React.PropTypes.number,
	unitType: React.PropTypes.string,
	unit: React.PropTypes.string,
	def: React.PropTypes.string,
	example: React.PropTypes.number,
};

export default Container.create(CodeFileController, {pure: false});
