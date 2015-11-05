import React from "react";
import Draggable from "react-draggable";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import ExampleView from "./ExampleView";
import DiscussionsList from "./DiscussionsList";
import DiscussionView from "./DiscussionView";
import MarkdownTextarea from "../components/MarkdownTextarea"; // FIXME
import hotLink from "./util/hotLink";

export default class DefPopup extends Component {
	constructor(props) {
		super(props);
		this.state = {
			viewAllDiscussions: false,
			viewDiscussion: null,
			newDiscussion: false,
			creatingDiscussion: false,
		};
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.examplesGeneration = props.examples.generation;
	}

	_createDiscussion() {
		this.setState({creatingDiscussion: true});
		let title = this.refs.titleText.value;
		let body = this.refs.bodyText.value();
		Dispatcher.dispatch(new DefActions.CreateDiscussion(this.state.def.URL, title, body, (d) => {
			this.setState({creatingDiscussion: false, viewDiscussion: d});
		}));
	}

	_renderBody() {
		let def = this.state.def;

		if (this.state.viewAllDiscussions) {
			return (
				<div className="discussions-list discussions">
					<div className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
					<div className="container">
						<div className="padded-form">
							<DiscussionsList
								discussions={this.state.discussions}
								onViewDiscussion={(d) => { this.setState({viewAllDiscussions: false, viewDiscussion: d}); }} />
						</div>
					</div>
					<footer>
						<a ref="createBtn" onClick={() => { this.setState({viewAllDiscussions: false, newDiscussion: true}); }}><i className="fa fa-comment" /> New</a>
					</footer>
				</div>
			);
		}

		if (this.state.viewDiscussion) {
			return (
				<div className="discussion-thread discussions">
					<DiscussionView discussion={this.state.viewDiscussion} def={def} />
					<footer>
						<a ref="listBtn" onClick={() => { this.setState({viewAllDiscussions: true, viewDiscussion: null}); }}><i className="fa fa-eye" /> View all</a>
						<a href="#add-discussion-comment"><i className="fa fa-plus" /> Reply</a>
						<a ref="createBtn" onClick={() => { this.setState({viewDiscussion: null, newDiscussion: true}); }}><i className="fa fa-comment" /> New</a>
					</footer>
				</div>
			);
		}

		if (this.state.newDiscussion) {
			return (
				<div className="discussion-create">
					<div className="form">
						<h1>Create a discussion</h1>
						<p>You are starting a new discussion on <b className="backtick" dangerouslySetInnerHTML={def.QualifiedName} />.</p>
						<input type="text" ref="titleText" className="title" placeholder="Title" />
						<MarkdownTextarea ref="bodyText" className="body" placeholder="Description" />
						<div className="buttons pull-right">
							<button ref="createBtn" className={`btn btn-sgblue ${this.state.creatingDiscussion ? "disabled" : ""}`} onClick={!this.state.creatingDiscussion && (() => { this._createDiscussion(); })}>Create</button>
							<button ref="cancelBtn" className={`btn btn-default ${this.state.creatingDiscussion ? "disabled" : ""}`} onClick={!this.state.creatingDiscussion && (() => { this.setState({newDiscussion: false}); })}>Cancel</button>
						</div>
					</div>
				</div>
			);
		}

		let topDiscussions = this.state.discussions && this.state.discussions.slice().sort((a, b) => {
			let d = b.Comments.length - a.Comments.length;
			if (d !== 0) {
				return d;
			}
			return b.ID - a.ID;
		}).slice(0, 4);

		return (
			<div>
				<section className="docHTML">
					<div className="header">
						<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
					</div>
					<section className="doc">
						{def.Found ? <span dangerouslySetInnerHTML={def.Data.DocHTML} /> : <span>Definition of <span dangerouslySetInnerHTML={def.QualifiedName} /> is not available.</span>}
					</section>
				</section>

				<ExampleView defURL={def.URL} examples={this.state.examples} highlightedDef={this.state.highlightedDef} />

				{this.state.discussions &&
					<div className="code-discussions">
						{this.state.discussions.length === 0 ? (
							<div className="no-discussions"><a ref="createBtn" onClick={() => { this.setState({newDiscussion: true}); }}><i className="octicon octicon-plus" /> Start a code discussion</a></div>
						) : (
							<div className="contents">
								<DiscussionsList
									discussions={topDiscussions}
									onViewDiscussion={(d) => { this.setState({viewDiscussion: d}); }}
									small={true} />
								<footer>
									<a ref="listBtn" onClick={() => { this.setState({viewAllDiscussions: true}); }}><i className="fa fa-eye" /> View all</a>
									<a ref="createBtn" onClick={() => { this.setState({newDiscussion: true}); }}><i className="fa fa-comment" /> New</a>
								</footer>
							</div>
						)}
					</div>
				}
			</div>
		);
	}

	render() {
		let def = this.state.def;
		return (
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							{(this.state.viewAllDiscussions || this.state.viewDiscussion || this.state.newDiscussion) &&
								<a key="back-to-main" className="btn btn-toolbar btn-default" onClick={() => { this.setState({viewAllDiscussions: false, viewDiscussion: null, newDiscussion: false}); }}>
									<span className="octicon octicon-arrow-left" /> Back
								</a>
							}

							{def.Found &&
								<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={hotLink}>Go to definition</a>
							}

							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</header>

						{this._renderBody()}
					</div>
				</div>
			</Draggable>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
	discussions: React.PropTypes.array,
};
