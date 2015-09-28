var React = require("react");
var moment = require("moment");

var MarkdownTextarea = require("./MarkdownTextarea");
var MarkdownView = require("./MarkdownView");
var DiscussionModel = require("../stores/models/DiscussionModel");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");

/**
 * @description DiscussionView displays all details about a discussion.
 */
var DiscussionView = React.createClass({

	propTypes: {
		// defName is the qualified name of the definition that will
		// be displayed in the UI.
		defName: React.PropTypes.object.isRequired,

		// defKey is the DefKey URL form of the definition that this
		// discussion belongs too.
		defKey: React.PropTypes.string.isRequired,

		// onList will be called if the user requests to list all discussions.
		onList: React.PropTypes.func,

		// onCreate is the callback called when the user requests to create a
		// new discussion.
		onCreate: React.PropTypes.func,

		// onComment is the callback called when the user requests to add a
		// comment.
		onComment: React.PropTypes.func,

		// model is an object containing information about the discussion to be
		// shown.
		model: React.PropTypes.instanceOf(DiscussionModel).isRequired,
	},

	mixins: [ModelPropWatcherMixin],

	getDefaultProps() {
		return {
			onList: () => {},
			onCreate: () => {},
			onComment: () => {},
		};
	},

	_onComment() {
		this.props.onComment(this.state.ID, this.refs.commentTextarea.value());
	},

	render() {
		var state = this.state;

		return (
			<div className="discussion-thread discussions">
				<div className="container">
					<div className="padded-form">
						<header>
							<h1>
								<div className="contents">
									{state.Title}<span className="id">{" #"+state.ID}</span>
								</div>
							</h1>
							<div className="stats">
								<span className="octicon octicon-comment-discussion" />{" "+state.Comments.length+" "}
								<span className="octicon octicon-star" />{" "+state.Ratings.length}
							</div>
							<div className="subtitle">
								<span className="author"><a>{"@"+state.Author.Login}</a></span>
								<span className="date">{" "+moment(state.CreatedAt).fromNow()}</span>
								<span className="subject"> on <b className="backtick" dangerouslySetInnerHTML={this.props.defName} /></span>
							</div>
						</header>
						{state.Description ? <main className="body">{state.Description}</main> : null}
						{state.Comments.length > 0 ? (
							<ul className="thread-comments">
								{state.Comments.map(c => (
									<li className="thread-comment" key={"discussion-view-comment-"+c.ID}>
										<div className="signature">
											<a>{"@"+c.Author.Login}</a> replied <i>{moment(c.CreatedAt).fromNow()}</i>
										</div>
										<MarkdownView content={c.Body} />
									</li>
								))}
							</ul>
						) : null}
					</div>
					<div className="add-comment">
						<div className="padder pull-left">
							<MarkdownTextarea className="thread-comment-add" ref="commentTextarea" />
							<a ref="commentBtn" id="add-discussion-comment" onClick={this._onComment} className="btn btn-sgblue pull-right">Comment</a>
						</div>
					</div>
				</div>
				<footer>
					<a ref="listBtn" onClick={this.props.onList.bind(this, this.props.defKey)}><i className="fa fa-eye" /> View all</a>
					<a href="#add-discussion-comment"><i className="fa fa-plus" /> Reply</a>
					<a ref="createBtn" onClick={this.props.onCreate}><i className="fa fa-comment" /> New</a>
				</footer>
			</div>
		);
	},
});

module.exports = DiscussionView;
