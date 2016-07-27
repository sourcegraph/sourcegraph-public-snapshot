import * as React from "react";

class Component extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
	}

	componentWillMount() {
		this._updateState(Object.assign({}, this.state), this.props, this.context);
	}

	componentWillReceiveProps(nextProps, nextContext) {
		this._updateState(Object.assign({}, this.state), nextProps, nextContext);
	}

	shouldComponentUpdate(nextProps, nextState, nextContext) {
		let keys = Object.keys(nextState);
		if (Object.keys(this.state).length !== keys.length) {
			return true;
		}
		for (let i = 0; i < keys.length; i++) {
			let k = keys[i];
			if (nextState[k] !== this.state[k]) {
				return true;
			}
		}
		return false;
	}

	setState(patch, callback) {
		this._updateState(Object.assign({}, this.state, patch), this.props, this.context, callback);
	}

	_updateState(newState, props, context, callback) {
		this._checkForUndefined(props, "Property");
		if (context) this._checkForUndefined(context, "Context");
		this.reconcileState(newState, props, context);
		this._checkForUndefined(newState, "State");
		if (this.onStateTransition) {
			this.onStateTransition(this.state, newState);
		}
		super.setState(newState, callback);
	}

	_checkForUndefined(obj, type) {
		if (process.env.NODE_ENV === "production") { return; }
		let keys = Object.keys(obj);
		for (let i = 0; i < keys.length; i++) {
			if (obj[keys[i]] === undefined) { // eslint-disable-line no-undefined
				throw new Error(`Invariant Violation: ${type} '${keys[i]}' of ${this.constructor.name} has value 'undefined'.`);
			}
		}
	}

	reconcileState(state, props, context) {
		// override
	}
}

export default Component;
