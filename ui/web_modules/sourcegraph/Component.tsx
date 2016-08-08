import * as React from "react";

class Component<P, S> extends React.Component<P, S> {
	constructor(props: P) {
		super(props);
		this.state = {} as S;
	}

	componentWillMount(): void {
		this._updateState(Object.assign({}, this.state), this.props, this.context);
	}

	componentWillReceiveProps(nextProps: P, nextContext: any): void {
		this._updateState(Object.assign({}, this.state), nextProps, nextContext);
	}

	shouldComponentUpdate(nextProps: P, nextState: S, nextContext: any): boolean {
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

	setState(patch: S | ((prevState: S, props: P) => S), callback?: () => any): void {
		if (patch instanceof Function) {
			throw new Error("setState with function parameter not supported");
		} else {
			this._updateState(Object.assign({} as S, this.state, patch), this.props, this.context, callback);
		}
	}

	_updateState(newState: S, props: P, context: any, callback?: () => any): void {
		this._checkForUndefined(props, "Property");
		if (context) { this._checkForUndefined(context, "Context"); }
		this.reconcileState(newState, props, context);
		this._checkForUndefined(newState, "State");
		this.onStateTransition(this.state, newState);
		super.setState(newState, callback);
	}

	_checkForUndefined(obj: any, type: string): void {
		if (process.env.NODE_ENV === "production") { return; }
		let keys = Object.keys(obj);
		for (let i = 0; i < keys.length; i++) {
			if (obj[keys[i]] === undefined) { // eslint-disable-line no-undefined
				throw new Error(`Invariant Violation: ${type} '${keys[i]}' of ${this.constructor.name} has value 'undefined'.`);
			}
		}
	}

	reconcileState(state: S, props: P, context: any): void {
		// override
	}

	onStateTransition(prevState: S, nextState: S): void {
		// override
	}
}

export default Component;
