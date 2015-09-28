var React = require("react");
var classNames = require("classnames");
var cookie = require("react-cookie");
var hashCode = require("../hashCode");

/**
 * @description AlertView displays a text alert. It may be configured so that it is closable, and optionally
 * so that it saves its state in a cookie - meaning that once closed, the exact same message will not be displayed
 * again.
 */
var AlertView = React.createClass({
	propTypes: {
		// content holds the body to be displayed as the alert.
		content: React.PropTypes.string,

		// html holds raw HTML for the body of the alert, used if a content
		// string is not available. Note that this can be dangerous -- you
		// should never pass any user-inputted text here or else you are opening
		// a XSS vulnerability. Use this field for hand-written HTML only!
		html: React.PropTypes.string,

		// icon is the font-awesome icon to use for the alert. By default this
		// is "fa-warning".
		icon: React.PropTypes.string,

		// closeable defaults to false. If true, it will allow
		// the alert to be permanently closed.
		closeable: React.PropTypes.bool,

		// hasCookie tells the component to permanently maintain the closed
		// state in a cookie.
		hasCookie: React.PropTypes.bool,

		// label, if specified, will identify a group of cookies which if closed
		// with 'hasCookie' enabled, will not be displayed again.
		label: React.PropTypes.string,
	},

	getInitialState() {
		return {
			closed: this.props.hasCookie ? Boolean(cookie.load(this._cookieCode())) : false,
		};
	},

	/**
	 * @description Handler called when the close button is clicked.
	 * @returns {void}
	 * @private
	 */
	_close() {
		if (this.props.hasCookie) {
			cookie.save(this._cookieCode(), "true");
		}

		this.setState({closed: true});
	},

	/**
	 * @description Returns the name of the cookie that stores that state of this component.
	 * @returns {void}
	 * @private
	 */
	_cookieCode() {
		return "alert-view-closed-" + hashCode(this.props.label || this.props.content || this.props.html);
	},

	render() {
		var cx = classNames({
			"alert-view": true,
			"closed": this.state.closed,
		});

		var text;
		if (this.props.html) {
			text = <td className="text" dangerouslySetInnerHTML={{__html: this.props.html}}></td>;
		} else {
			text = <td className="text">{this.props.content}</td>;
		}

		var icon = this.props.icon ? this.props.icon : "fa-warning";

		return (
			<div className={cx}>
				{this.props.closeable ? <div className="btn-close" onClick={this._close}>×</div> : null}
				<table className="over-threshold-warning">
					<tbody>
						<td className="icon"><i className={"fa fa-icon " + icon} /></td>
						{text}
					</tbody>
				</table>
			</div>
		);
	},
});

module.exports = AlertView;
