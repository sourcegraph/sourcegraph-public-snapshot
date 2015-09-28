var React = require("react");

var ContextMenuModel = require("../stores/models/ContextMenuModel");
var ModelPropWatcherMixin = require("./mixins/ModelPropWatcherMixin");

var ContextMenu = React.createClass({
	propTypes: {
		model: React.PropTypes.instanceOf(ContextMenuModel).isRequired,
		onClick: React.PropTypes.func,
	},

	mixins: [ModelPropWatcherMixin],

	_onClick(data, evt) {
		if (typeof this.props.onClick === "function") {
			this.props.onClick(data, evt);
		}
	},

	render() {
		if (!this.state.options.length || this.state.closed) {
			return null;
		}
		return (
			<div className="context-menu"
				style={{
					top: this.state.position.top,
					left: this.state.position.left,
				}}>
				<ul>
					{this.state.options.map(opt =>
						// opt.label comes from _onReceivedMenuOptions in
						// CodeFileStore.js, or rather it is QualifiedName from
						// ui/def_list.go which means that this is safe to do
						// because it is escaped properly there. Also we don't use
						// __html here because it is of type pbtypes.HTML which is
						// marshaled into an __html type (not a string).
						<li dangerouslySetInnerHTML={opt.label}
							key={opt.data.URL}
							onClick={this._onClick.bind(this, opt.data)} />
					)}
				</ul>
			</div>
		);
	},
});

module.exports = ContextMenu;
