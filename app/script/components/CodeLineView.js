var React = require("react");
var ReactDOM = require("react-dom");
var globals = require("../globals");
var $ = require("jquery");
var classNames = require("classnames");

var CodeFileActions = require("../actions/CodeFileActions");
var CodeToken = require("./CodeTokenView");
var CodeLineModel = require("../stores/models/CodeLineModel");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");

/**
 * @description CodeLineView displays a line of code. Optionally, line number
 * can be hidden or displayed.
 */
var CodeLineView = React.createClass({

	// The properties of this element are not applied directly. They are passed down
	// from the containing CodeView.
	propTypes: {
		// The loading property is passed down from the CodeView.
		loading: React.PropTypes.bool,

		// The model of the token to be displayed.
		model: React.PropTypes.instanceOf(CodeLineModel).isRequired,

		// Whether to display line numbers.
		lineNumbers: React.PropTypes.bool,

		// The function to be called on click. It will receive as arguments the
		// CodeTokenModel that was clicked and the event. Default is automatically
		// prevented.
		onTokenClick: React.PropTypes.func,

		// The function to be called on 'mouseenter'. It will receive as arguments the
		// CodeTokenModel and the event. Default is automatically prevented.
		onTokenFocus: React.PropTypes.func,

		// The function to be called on 'mouseleave'. It will receive as arguments the
		// CodeTokenModel and the event. Default is automatically prevented.
		onTokenBlur: React.PropTypes.func,

		// onComment is a function that will be triggered if the comment button is visible
		// and clicked.
		onComment: React.PropTypes.func,

		// allowComments will display the comment '+' button next to each row if the line
		// shows a diff.
		allowComments: React.PropTypes.bool,
	},

	mixins: [ModelPropWatcherMixin],

	componentDidMount() {
		if (this.isMounted()) this.props.model.__node = $(ReactDOM.findDOMNode(this));
	},

	/**
	 * @description Triggered when the line number is clicked on a line.
	 * @param {Event} e - Event
	 * @returns {void}
	 * @private
	 */
	_clickLine(e) {
		CodeFileActions.selectLines(this.state.number, e.shiftKey);
	},

	_onCommentClick(e) {
		if (typeof this.props.onComment === "function") {
			this.props.onComment(this.props.model, e);
		}
	},

	render() {
		var classes = classNames({
			"line": true,
			"main-byte-range": this.state.highlight,
			"new-line": this.state.prefix === "+",
			"old-line": this.state.prefix === "-",
		}) + (this.state.extraClass ? ` ${this.state.extraClass}` : "");

		return (
			<tr className={classes} data-start={this.state.start} data-end={this.state.end} style={this.props.style}>
				{this.props.lineNumbers !== false ? (
					<td className="line-number" data-line={this.state.number} onClick={this._clickLine}></td>
				) : null}

				{typeof this.state.lineNumberBase !== "undefined" ? (
					<td className="line-number" data-line={this.state.lineNumberBase}></td>
				) : null}

				{typeof this.state.lineNumberHead !== "undefined" ? (
					<td className="line-number" data-line={this.state.lineNumberHead}>
						{this.props.allowComments && this.state.allowComments !== false ? (
							<a className="btn-inline-comment" onClick={this._onCommentClick}>
								<span className="octicon octicon-plus"></span>
							</a>
						) : null}
					</td>
				) : null}

				<td className="line-content">
					{typeof this.state.prefix === "string" ? (
						<span className="prefix">{this.state.prefix}</span>
					) : null}

					{this.state.tokens.length ? this.state.tokens.map(token => {
						var xClass = token.get("extraClass") || "";
						switch (token.get("type")) {
						case globals.TokenType.STRING:
							return <span key={token.cid} className={xClass}>{token.get("html")}</span>;

						case globals.TokenType.SPAN:
							return <span key={token.cid} className={`${token.get("syntax")} ${xClass}`}>{token.get("html")}</span>;

						default:
							return <CodeToken {...this.props} key={token.cid} model={token} />;
						}
					}) : " "}
				</td>
			</tr>
		);
	},
});

module.exports = CodeLineView;
