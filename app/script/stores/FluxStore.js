var Backbone = require("backbone");
var globals = require("../globals");

/**
 * @description FluxStore is a factory for stores that facilitate FLUX architectures.
 * It is a wrapper on Backbone.Model which allows the usage of an actions map that
 * registers action types to functions on a set dispatcher.
 *
 * Example usage:
 *
 *     var MyStore = FluxStore({
 *         dispatcher: MyDispatcher,
 *
 *         actions: {
 *             "ACTION_NAME": "_onActionName",
 *             "OTHER_ACTION": "_onOtherAction"
 *         },
 *
 *         _onActionName(action, source) {
 *              // respond to action
 *         },
 *
 *         _onOtherAction(action, source) {
 *             // respond to action
 *         }
 *     });
 *
 * Callback functions get passed the action payload, along with the source.
 *
 * @param {Object} body - The body to extend Backbone.Model with. It is safe to use all
 * of the Backbone.Model methods, including initialize, which is called *after* FluxStore's
 * initialize.
 * @returns {Backbone.Model} The resulting model.
 */
module.exports = function(body) {
	var initializeSuper = body.initialize || function noop() {},
		destroySuper = body.destroy || function noop() {};

	body.initialize = function() {
		if (typeof this.actions === "object") {
			// Validate actions
			Object.keys(this.actions).forEach(action => {
				if (!globals.Actions.hasOwnProperty(action)) {
					console.warn(`Action ${action} not found in (globals.Actions) definitions`);
				}

				if (typeof this[this.actions[action]] !== "function") {
					console.warn(`Action ${action} is bound to inexistent callback: ${this.actions[action]}`);
				}
			});
			// Register with set dispatcher
			this.dispatchToken = this.dispatcher.register(payload => {
				var type = payload.action.type,
					isRegistered = this.actions.hasOwnProperty(type),
					callbackExists = typeof this[this.actions[type]] === "function";

				if (isRegistered && callbackExists) {
					this[this.actions[type]](payload.action, payload.source);
				}
			});
		}

		Reflect.apply(initializeSuper, this);
	};

	body.destroy = function() {
		this.dispatcher.unregister(this.dispatchToken);
		Reflect.apply(destroySuper, this);
	};

	return new (Backbone.Model.extend(body))();
};
