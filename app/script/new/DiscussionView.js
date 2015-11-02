import React from "react";
import moment from "moment";

import Component from "./Component";
import MarkdownView from "../components/MarkdownView"; // FIXME
import MarkdownTextarea from "../components/MarkdownTextarea"; // FIXME
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";

export default class DiscussionView extends Component {
	constructor(props) {
		super(props);
		this.state = {
			creatingComment: false,
		};
	}

	reconcileState(state, props) {
		state.discussion = props.discussion;
		state.def = props.def;
	}

	_createComment() {
		if (!this.refs.commentTextarea) { return; } // FIXME workaround for tests
		this.setState({creatingComment: true});
		let body = this.refs.commentTextarea.value();
		Dispatcher.dispatch(new DefActions.CreateDiscussionComment(this.state.def.URL, this.state.discussion.ID, body, () => {
			this.refs.commentTextarea.value("");
			this.setState({creatingComment: false});
		}));
	}

	render() {
		let d = this.state.discussion;
		return (
			<div className="container">
				<div className="padded-form">
					<header>
						<h1>
							<div className="contents">
								{d.Title}<a className="id">{` #${d.ID}`}</a>
							</div>
						</h1>
						<div className="stats">
							<span className="octicon octicon-comment-discussion" /> {d.Comments.length}
						</div>
						<div className="subtitle">
							<span className="author"><a>@{d.Author.Login}</a></span>
							<span className="date"> {moment(d.CreatedAt).fromNow()}</span>
							<span className="subject"> on <b className="backtick" dangerouslySetInnerHTML={this.state.def.QualifiedName} /></span>
						</div>
					</header>
					{d.Description ? <main className="body">{d.Description}</main> : null}
					<ul className="thread-comments">
						{d.Comments && d.Comments.map(c => (
							<li className="thread-comment" key={c.ID}>
								<div className="signature">
									<a>@{c.Author.Login}</a> replied <i>{moment(c.CreatedAt).fromNow()}</i>
								</div>
								<MarkdownView content={c.Body} />
							</li>
						))}
					</ul>
				</div>
				<div className="add-comment">
					<div className="padder pull-left">
						<MarkdownTextarea className="thread-comment-add" ref="commentTextarea" />
						<a className={`btn btn-sgblue pull-right ${this.state.creatingComment ? "disabled" : ""}`} onClick={!this.state.creatingComment && (() => { this._createComment(); })}>Comment</a>
					</div>
				</div>
			</div>
		);
	}
}

DiscussionView.propTypes = {
	discussion: React.PropTypes.object,
	def: React.PropTypes.object,
};
