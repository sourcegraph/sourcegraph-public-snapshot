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
