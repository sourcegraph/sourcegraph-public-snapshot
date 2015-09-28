var React = require("react");
var $ = require("jquery");
var classNames = require("classnames");
var globals = require("../globals");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var CodeTokenModel = require("../stores/models/CodeTokenModel");

var CodeTokenView = React.createClass({

	// The properties of this element are not applied directly. They are passed down
	// via the CodeLineView from the containing CodeView.
	propTypes: {
		// The loading property is passed down from the CodeView via CodeLineView.
		// If set, there will be no response to user actions on this token.
		loading: React.PropTypes.bool,

		// The model of the token to be displayed.
		model: React.PropTypes.instanceOf(CodeTokenModel).isRequired,

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
	},

	mixins: [ModelPropWatcherMixin],

	componentDidMount() {
		if (!this.isMounted()) return;

		var el = this.getDOMNode();

		if (typeof this.props.onTokenFocus === "function") {
			el.addEventListener("mouseenter", this._onFocus);
		}

		if (typeof this.props.onTokenBlur === "function") {
			el.addEventListener("mouseleave", this._onBlur);
		}

		this.props.model.__node = $(el);
	},

	componentWillUnmount() {
		if (!this.isMounted()) return;

		var el = this.getDOMNode();
		el.removeEventListener("mouseenter", this._onFocus);
		el.removeEventListener("mouseleave", this._onBlur);
	},

	/**
	 * @description Called when token is focused using the 'mouseenter' event.
	 * @param {Event} event - Event
	 * @returns {void}
	 * @private
	 */
	_onFocus(event) {
		if (this.props.loading) return;
		this.props.onTokenFocus(this.props.model, event);
	},

	/**
	 * @description Called when token focus is lost on 'mouseleave' event.
	 * @param {Event} event - Event
	 * @returns {void}
	 * @private
	 */
	_onBlur(event) {
		if (this.props.loading) return;
		this.props.onTokenBlur(this.props.model, event);
	},

	/**
	 * @description Triggered on click event. Noop if component is loading. Otherwise,
	 * prop callback is called if set.
	 * @param {Event} e - Event
	 * @returns {void}
	 * @private
	 */
	_onClick(e) {
		if (this.props.loading) return;
		if (typeof this.props.onTokenClick === "function") {
			this.props.onTokenClick(this.props.model, e);
			e.preventDefault();
		}
	},

	render() {
		var t = this.state.type;
		var classes = classNames({
			"ref": t === globals.TokenType.REF || t === globals.TokenType.DEF,
			"def": t === globals.TokenType.DEF,
			"highlight-secondary": !this.state.selected && this.state.highlighted,
			"highlight-primary": this.state.selected,
		});

		return (
			<a href={this.state.url} className={classes} onClick={this._onClick}>
				<span className={this.state.syntax + " " + this.state.extraClass}>{this.state.html}</span>
			</a>
		);
	},
});

module.exports = CodeTokenView;
