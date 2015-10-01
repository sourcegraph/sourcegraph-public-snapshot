var React = require("react");
var moment = require("moment");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var DiscussionCollection = require("../stores/collections/DiscussionCollection");

/**
 * @description Displays a list of discussions. See PropTypes for configuration.
 */
var DiscussionList = React.createClass({

	propTypes: {
		// defName is the qualified name of the definition that will
		// be displayed in the UI.
		defName: React.PropTypes.object.isRequired,

		// defKey is the DefKey URL form of the definition that this
		// discussion belongs too.
		defKey: React.PropTypes.string.isRequired,

		// model is the list of discussions to display.
		model: React.PropTypes.instanceOf(DiscussionCollection).isRequired,

		// onCreate is the callback called when the user requests to create a
		// new discussion.
		onCreate: React.PropTypes.func,

		// onClick is the click event callback for elements in the list.
		// The callback receives the DefKey ID (in URL form) and the clicked
		// discussion as parameters.
		onClick: React.PropTypes.func,
	},

	mixins: [ModelPropWatcherMixin],

	getDefaultProps() {
		return {
			onCreate: () => {},
			onClick: () => {},
		};
	},

	_renderDiscussion(model) {
		var d = model.attributes;
		return (
			<li onClick={this.props.onClick.bind(this, this.props.defKey, d)} key={`list-item-${d.ID}`}>
				<header>
					<h1>
						<div className="contents">
							<a>{d.Title}</a><span className="id">{` #${d.ID}`}</span>
						</div>
					</h1>
					<div className="stats">
						<span className="octicon octicon-comment-discussion" />{` ${d.Comments.length} `}
					</div>
					<div className="subtitle">
						<span className="author"><a>{`@${d.Author.Login}`}</a></span>
						<span className="date">{` ${moment(d.CreatedAt).fromNow()}`}</span>
					</div>
				</header>
				<p className="body">{d.Description.slice(0, 250)+(d.Description.length > 250 ? "..." : "")}</p>
			</li>
		);
	},

	render() {
		var list = this.state.models;

		return (
			<div className="discussions-list discussions">
				<div className="qualified-name" dangerouslySetInnerHTML={this.props.defName} />
				<div className="container">
					<div className="padded-form">
						<ul className="discussions-list">
							{list.map(this._renderDiscussion)}
						</ul>
					</div>
				</div>
				<footer>
					<a ref="createBtn" onClick={this.props.onCreate}><i className="fa fa-comment" /> New</a>
				</footer>
			</div>
		);
	},
});

module.exports = DiscussionList;
