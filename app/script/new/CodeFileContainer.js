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
		this._requestData();
	}

	componentWillReceiveProps(nextProps) {
		this._requestData();
	}

	_requestData() {
		setTimeout(() => {
			Dispatcher.dispatch(new CodeActions.WantFile(this.props.repo, this.props.rev, this.props.tree));
			if (this.props.selectedDef) {
				Dispatcher.dispatch(new DefActions.WantDef(this.props.selectedDef));
			}
		}, 0);
	}

	static getStores() {
		return [CodeStore, DefStore];
	}

	static calculateState(prevState) {
		return {
			files: CodeStore.files,
			defs: DefStore.defs,
			examples: DefStore.examples,
			highlightedDef: DefStore.highlightedDef,
		};
	}

	render() {
		let file = this.state.files.get(this.props.repo, this.props.rev, this.props.tree);
		if (!file) {
			return null;
		}
		let def = this.props.selectedDef && this.state.defs.get(this.props.selectedDef);
		return (
			<div>
				<div className="code-view-react">
					<CodeListing
						lines={file.Entry.SourceCode.Lines}
						lineNumbers={true}
						selectedDef={this.props.selectedDef}
						highlightedDef={this.state.highlightedDef} />
				</div>
				{def &&
					<DefPopup
						def={def}
						examples={this.state.examples}
						highlightedDef={this.state.highlightedDef} />
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

export default Container.create(CodeFileContainer, {pure: false});
