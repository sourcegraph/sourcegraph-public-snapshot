import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/modal.css";

class Modal extends Component {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_onClick(e) {
		if (e.target === this.refs.modal_container) {
			if (this.state.onDismiss) this.state.onDismiss();
		}
	}

	render() {
		return (
			<div ref="modal_container"
				styleName={this.state.shown ? "container" : "hidden"}
				onClick={this._onClick}>
					{this.state.children}
			</div>
		);
	}
}

Modal.propTypes = {
	shown: React.PropTypes.bool.isRequired,
	onDismiss: React.PropTypes.func,
};

Modal = CSSModules(Modal, styles);
export default Modal;

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, stateKey, children, onDismiss}, {router}) {
	const onDismiss2 = () => {
		router.replace({...location, state: {...location.state, [stateKey]: false}});
		if (onDismiss) onDismiss();
	};
	return (
		<Modal shown={location.state && Boolean(location.state[stateKey])}
			onDismiss={onDismiss2}>
			{children}
		</Modal>
	);
}
LocationStateModal.propTypes = {
	location: React.PropTypes.object.isRequired,

	// stateKey is the name of the key on the location's state that this
	// StateToggleLink component toggles.
	stateKey: React.PropTypes.string.isRequired,

	onDismiss: React.PropTypes.func,
};
LocationStateModal.contextTypes = {
	router: React.PropTypes.object.isRequired,
};
