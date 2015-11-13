import React from "react";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import MarkdownTextarea from "../components/MarkdownTextarea"; // FIXME
import * as IssueActions from "./IssueActions";
import "./IssueBackend";

export default class IssueForm extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_createIssue() {
		Dispatcher.asyncDispatch(new IssueActions.CreateIssue(
			this.state.repo,
			this.state.path,
			this.state.commitID,
			this.state.startLine,
			this.state.endLine,
			this.refs.title.value,
			this.refs.body.value(),
			this.state.onSubmit
		));
	}

	render() {
		return (
			<div className="inline-content">
				<p><b>Creating an issue on {this.state.path}:{this.state.startLine}-{this.state.endLine}</b></p>
				<div className="inline-form">
					<input ref="title" type="text" placeholder="Title" autoFocus="true"/>
					<MarkdownTextarea ref="body" placeholder="Description"/>
					<div className="actions">
						<button className="btn btn-success" tabIndex="0" onClick={() => { this._createIssue(); }}>Create Issue</button>
						<button className="btn btn-neutral" tabIndex="0" onClick={this.state.onCancel}>Cancel</button>
					</div>
				</div>
			</div>
		);
	}
}
