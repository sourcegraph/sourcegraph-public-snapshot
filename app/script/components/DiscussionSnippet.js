var React = require("react");
var globals = require("../globals");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var DiscussionCollection = require("../stores/collections/DiscussionCollection");

var DiscussionSnippet = React.createClass({

	propTypes: {
		// defKey is the DefKey URL form of the definition that this
		// discussion belongs too.
		defKey: React.PropTypes.string.isRequired,

		// onClick is called when a discussion is clicked.
		onClick: React.PropTypes.func.isRequired,

		// onList will be called if the user requests to list all discussions.
		onList: React.PropTypes.func,

		// onCreate is the callback called when the user requests to create a
		// new discussion.
		onCreate: React.PropTypes.func,

		// toolbar, when true, will show the toolbar with links to create a new
		// discussion or view all discussions.
		toolbar: React.PropTypes.bool,

		// model holds a collection of DiscussionModel's shown in the snippet.
		model: React.PropTypes.instanceOf(DiscussionCollection).isRequired,
	},

	mixins: [ModelPropWatcherMixin],

	_hasDiscussions() {
		var list = this.state.models;
		return typeof list !== "undefined" && list.length > 0;
	},

	_renderDiscussion(model) {
		var d = model.attributes;
		var url = this.props.defKey;

		return (
			<li className="discussion" key={`snippet-d-${d.ID}`}>
				<a onClick={e => this.props.onClick(url, d, e)} className="title truncate">{d.Title}</a>
				<div className="stats">
					<span className="octicon octicon-comment-discussion" /> {d.Comments.length}
					&nbsp;<span className="octicon octicon-star" /> {d.Ratings.length}
				</div>
				<p className="body truncate">{d.Description}</p>
			</li>
		);
	},

	render() {
		return (
			<div className="code-discussions">
				{!this._hasDiscussions() ? (
					<div className="no-discussions">There are no discussions on this item, <a ref="createBtn" onClick={this.props.onCreate}>click here</a> to start one.</div>
				) : (
					<div className="contents">
						<ul className="list">
							{this.state.models.slice(0, globals.DiscussionSnippetEntries).map(this._renderDiscussion)}
						</ul>
						{this.props.toolbar ? (
							<footer>
								<a ref="listBtn" onClick={e => this.props.onList(this.props.defKey, e)}><i className="fa fa-eye" /> View all</a>
								<a ref="createBtn" onClick={this.props.onCreate}><i className="fa fa-comment" /> New</a>
							</footer>
						) : null}
					</div>
				)}
			</div>
		);
	},
});

module.exports = DiscussionSnippet;
