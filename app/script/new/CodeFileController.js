import React from "react";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import CodeStore from "./CodeStore";
import CodeListing from "./CodeListing";
import "./CodeBackend";

class CodeFileController extends React.Component {
	constructor(props, context) {
		super(props, context);
		this._onStoreChange = this._onStoreChange.bind(this);

		this.state = {
			files: CodeStore.files,
		};
	}

	componentWillMount() {
		Dispatcher.dispatch(new CodeActions.WantFile(this.props.repo, this.props.rev, this.props.tree));
	}

	componentDidMount() {
		CodeStore.addListener(this._onStoreChange);
	}

	componentWillUnmount() {
		CodeStore.removeListener(this._onStoreChange);
	}

	_onStoreChange() {
		this.setState({
			files: CodeStore.files,
		});
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

export default CodeFileController;
