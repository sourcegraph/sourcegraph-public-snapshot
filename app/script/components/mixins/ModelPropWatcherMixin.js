/**
 * @description This mixin links the host component's state to a Backbone model or collection
 * found in its props as this.props.model. If the property is not found, an error
 * is thrown.
 *
 * TODO(gbbr): Better distinction between 'model' and 'collection' via different view props.
 */
module.exports = {
	getInitialState() {
		if (!this.props.model) throw new Error("Need to set valid model attribute for ModelPropWatcherMixin.");
		return this.props.model.attributes || {models: this.props.model.models};
	},

	componentDidMount() {
		this.props.model.on("add remove change",
			() => this.setState(this.props.model.attributes || {models: this.props.model.models})
		);
	},

	componentWillUnmount() {
		this.props.model.off();
	},
};
