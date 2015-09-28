var React = require("react");
var router = require("../routing/router");

var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var DiffStatScale = require("./DiffStatScale");
var Hunk = require("./HunkView");

var FileDiffView = React.createClass({

	propTypes: {
		// Token event callback.
		// The function to be called on click. It will receive as arguments the
		// CodeTokenModel that was clicked and the event. Default is automatically
		// prevented.
		onTokenClick: React.PropTypes.func,

		// Token event callback.
		// The function to be called on 'mouseenter'. It will receive as arguments the
		// CodeTokenModel, event and file diff. Default is automatically prevented.
		onTokenFocus: React.PropTypes.func,

		// Token event callback.
		// The function to be called on 'mouseleave'. It will receive as arguments the
		// CodeTokenModel, event and file diff. Default is automatically prevented.
		onTokenBlur: React.PropTypes.func,

		// Function is called when the expand hunk is pressed in either direction.
		// It will call the function using parameters: hunk, direction and event.
		onExpandHunk: React.PropTypes.func,

		// allowComments will display the comment '+' button next to each line of code.
		allowComments: React.PropTypes.bool,

		// onCommentEdit is triggered when a comment is edited. It is passed the file diff,
		// hunk model, line model, comment model, the new body (string) and event.
		onCommentEdit: React.PropTypes.func,

		// onCommentDelete is triggered when a comment is deleted. It is passed the file diff,
		// hunk model, line model, comment model and event.
		onCommentDelete: React.PropTypes.func,

		// onCommentSubmit is triggered when a comment is submitted. It is passed
		// file diff model, hunk model, line model, body (string) and event.
		onCommentSubmit: React.PropTypes.func,
	},

	mixins: [ModelPropWatcherMixin],

	componentDidMount() {
		if (this.isMounted) this.props.model.__node = require("jquery")(this.getDOMNode());
	},

	_onTokenFocus(token, evt) {
		if (typeof this.props.onTokenFocus === "function") {
			this.props.onTokenFocus(token, evt, this.props.model);
		}
	},

	_onTokenBlur(token, evt) {
		if (typeof this.props.onTokenBlur === "function") {
			this.props.onTokenBlur(token, evt, this.props.model);
		}
	},

	_onTokenClick(token, evt) {
		if (typeof this.props.onTokenClick === "function") {
			this.props.onTokenClick(token, evt, this.props.model);
		}
	},

	_onCommentSubmit(hunk, line, body, evt) {
		if (typeof this.props.onCommentSubmit === "function") {
			this.props.onCommentSubmit(this.props.model, hunk, line, body, evt);
		}
	},

	_onCommentDelete(hunk, line, comment, evt) {
		if (typeof this.props.onCommentDelete === "function") {
			this.props.onCommentDelete(this.props.model, hunk, line, comment, evt);
		}
	},

	_onCommentEdit(hunk, line, comment, newBody, evt) {
		if (typeof this.props.onCommentEdit === "function") {
			this.props.onCommentEdit(this.props.model, hunk, line, comment, newBody, evt);
		}
	},

	render() {
		var baseUrl = router.fileURL(this.props.Delta.Base.URI, this.props.Delta.Base.CommitID, this.state.OrigName),
			newUrl = router.fileURL(this.props.Delta.Head.URI, this.props.Delta.Head.CommitID, this.state.NewName);

		var diffLinks = [];
		if (this.state.OrigName !== "/dev/null") {
			diffLinks.push(<a className="button btn btn-default btn-xs" href={baseUrl}>Original</a>);
		}
		if (this.state.NewName !== "/dev/null") {
			diffLinks.push(<a className="button btn btn-default btn-xs" href={newUrl}>New</a>);
		}

		return (
			<div className="file-diff">
				<header>
					<DiffStatScale Stat={this.state.Stats} />

					<span>{this.state.OrigName === "/dev/null" ? this.state.NewName : this.state.OrigName}</span>
					{this.state.NewName !== this.state.OrigName && this.state.OrigName !== "/dev/null" ? (
						<span> <i className="fa fa-long-arrow-right" /> {this.state.NewName}</span>
					) : null}

					<div className="btn-group pull-right">
						{diffLinks}
					</div>
				</header>

				{this.state.Hunks.map(
					hunk => <Hunk
						{...this.props}
						onTokenFocus={this._onTokenFocus}
						onTokenBlur={this._onTokenBlur}
						onTokenClick={this._onTokenClick}
						onCommentSubmit={this._onCommentSubmit}
						onCommentDelete={this._onCommentDelete}
						onCommentEdit={this._onCommentEdit}
						model={hunk}
						key={hunk.cid} />
				)}
			</div>
		);
	},
});

module.exports = FileDiffView;
