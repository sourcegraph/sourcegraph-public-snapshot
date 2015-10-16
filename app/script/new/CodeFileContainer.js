import React from "react";
import {Container} from "flux/utils";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";
import CodeStore from "./CodeStore";
import DefStore from "./DefStore";
import CodeListing from "./CodeListing";
import DefPopup from "./DefPopup";
import "./CodeBackend";
import "./DefBackend";

class CodeFileContainer extends React.Component {
	componentWillMount() {
		this._requestData(this.props);
	}

	componentWillReceiveProps(nextProps) {
		this._requestData(nextProps);
	}

	_requestData(props) {
		Dispatcher.dispatch(new CodeActions.WantFile(props.repo, props.rev, props.tree));
		if (props.selectedDef) {
			Dispatcher.dispatch(new DefActions.WantDef(props.selectedDef));
		}
	}

	static getStores() {
		return [CodeStore, DefStore];
	}

	static calculateState(prevState) {
		return {
			files: CodeStore.files,
			defs: DefStore.defs,
			highlightedDef: DefStore.highlightedDef,
		};
	}

	render() {
		let file = this.state.files.get(this.props.repo, this.props.rev, this.props.tree);
		if (!file) {
			return null;
		}
		let def = this.props.selectedDef && this.state.defs[this.props.selectedDef];
		return (
			<div>
				<CodeListing
					lines={file.Entry.SourceCode.Lines}
					selectedDef={this.props.selectedDef}
					highlightedDef={this.state.highlightedDef} />
				{def && <DefPopup def={def} />}
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

export default Container.create(CodeFileContainer, {pure: false});
