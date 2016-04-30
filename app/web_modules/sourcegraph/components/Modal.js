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

// dismissModal creates a function that dismisses the modal by setting
// the location state's modal property to null.
export function dismissModal(modalName, location, router) {
	return () => {
		if (location.state && location.state.modal !== modalName) {
			console.error(`location.state.modal is not ${modalName}, is:`, location.state.modal);
		}
		router.replace({...location, state: {...location.state, modal: null}});
	};
}

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, modalName, children, onDismiss}, {router}) {
	const onDismiss2 = () => {
		dismissModal(modalName, location, router)();
		if (onDismiss) onDismiss();
	};
	return (
		<Modal shown={location.state && location.state.modal === modalName}
			onDismiss={onDismiss2}>
			{children}
		</Modal>
	);
}
LocationStateModal.propTypes = {
	location: React.PropTypes.object.isRequired,

	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: React.PropTypes.string.isRequired,

	onDismiss: React.PropTypes.func,
};
LocationStateModal.contextTypes = {
	router: React.PropTypes.object.isRequired,
};
