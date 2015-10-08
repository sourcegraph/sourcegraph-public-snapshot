var React = require("react");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");
var $ = require("jquery");

var TokenPopoverView = React.createClass({

	mixins: [ModelPropWatcherMixin],

	/**
	 * @description Called on 'mousemove'. This function is only bound when popover
	 * is visible
	 * @param {event} evt - Mouse event
	 * @returns {void}
	 * @private
	 */
	_followMouse(evt) {
		if (this.isMounted()) {
			var x = evt.clientX, pw = 380; // popover width
			if (x > window.innerWidth-pw) x = window.innerWidth-pw;

			$(React.findDOMNode(this)).css({
				top: evt.clientY + 15,
				left: x + 15,
			});
		}
	},

	render() {
		var eventFn = this.state.visible ? "addEventListener" : "removeEventListener";
		document[eventFn]("mousemove", this._followMouse);

		return (
			<div className="token-popover"
				style={{
					display: this.state.visible ? "block" : "none",
					top: this.state.position.top,
					left: this.state.position.left,
				}}
				// This is OK because the body of the popover is a template
				// (def/popover.html), whose contents are sanitized by the app
				// (not untrusted federation root server) This happens in
				// util/handlerutil/repo.go by GetDefCommon primarily, but
				// also in app/def.go by serveDefPopover at a higher level.
				dangerouslySetInnerHTML={this.state.body} />
		);
	},
});

module.exports = TokenPopoverView;
