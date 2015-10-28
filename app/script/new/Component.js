import React from "react";

export default class Component extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
		this.updateState(this.state, props);
	}

	componentWillMount() {
		this._asyncRequestData();
	}

	componentWillReceiveProps(nextProps) {
		let newState = Object.assign({}, this.state);
		this.updateState(newState, nextProps);
		this.setState(newState, () => { this._asyncRequestData(); });
	}

	shouldComponentUpdate(nextProps, nextState) {
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

	patchState(patch) {
		let newState = Object.assign({}, this.state, patch);
		this.updateState(newState, this.props);
		this.setState(newState, () => { this._asyncRequestData(); });
	}

	_asyncRequestData() {
		setTimeout(() => {
			this.requestData();
		}, 0);
	}

	updateState(state, props) {
		// override
	}

	requestData() {
		// override
	}
}
