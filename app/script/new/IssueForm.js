import React from "react";

import Component from "./Component";
import MarkdownTextarea from "../components/MarkdownTextarea"; // FIXME

export default class IssueForm extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_createIssue() {
		console.log({
			repo: this.state.repo,
			commitID: this.state.commitID,
			path: this.state.path,
			startLine: this.state.startLine,
			endLine: this.state.endLine,
			title: this.refs.title.value,
			body: this.refs.body.value(),
		});
	}

	render() {
		return (
			<div className="inline-content">
				<p>Creating an issue on {this.state.path}:{this.state.startLine}-{this.state.endLine}</p>
				<div className="inline-form">
					<input ref="title" type="text" placeholder="Title" autoFocus="true"/>
					<MarkdownTextarea ref="body" placeholder="Description"/>
					<div className="actions">
						<button className="btn btn-success" tabIndex="0" onClick={() => { this._createIssue(); }}>Create Issue</button>
						<button className="btn btn-cancel" tabIndex="0" onClick={this.state.onCancel}>Cancel</button>
					</div>
				</div>
			</div>
		);
	}
}
