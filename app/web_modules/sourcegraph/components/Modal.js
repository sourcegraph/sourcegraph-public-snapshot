import React from "react";

import Component from "sourcegraph/Component";

import BaseStyles from "./styles/base.css";

class Modal extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return <div className={BaseStyles.modal_container} onClick={this.state.onClickOverlay}>{this.state.children}</div>;
	}
}

Modal.propTypes = {
	onClickOverlay: React.PropTypes.func,
};

export default Modal;
