import * as React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/modal.css";
import renderedOnBody from "sourcegraph/util/renderedOnBody";

class Modal extends React.Component {
	constructor(props) {
		super(props);
		this._onClick = this._onClick.bind(this);
		this._handleKeydown = this._handleKeydown.bind(this);
	}

	componentDidMount() {
		if (typeof document !== "undefined") {
			document.addEventListener("keydown", this._handleKeydown);

			// Prevent the page below the modal from scrolling.
			document.body.style.overflow = "hidden";
		}
	}

	componentWillUnmount() {
		if (typeof document !== "undefined") {
			document.removeEventListener("keydown", this._handleKeydown);
			document.body.style.overflow = "";
		}
	}

	_onClick(e) {
		if (e.target === this.refs.modal_container) {
			if (this.props.onDismiss) this.props.onDismiss();
		}
	}

	_handleKeydown(e: KeyboardEvent) {
		if (e.keyCode === 27 /* ESC */) {
			if (this.props.onDismiss) this.props.onDismiss();
		}
	}

	render() {
		return (
			<div ref="modal_container"
				styleName="container"
				onClick={this._onClick}>
					{this.props.children}
			</div>
		);
	}
}

Modal.propTypes = {
	onDismiss: React.PropTypes.func,
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
	]),
};

Modal = renderedOnBody(CSSModules(Modal, styles));
export default Modal;

// setLocationModalState shows or hides a modal by setting the location.state.modal
// property to modalName if shown is true and null otherwise.
export function setLocationModalState(router: any, location: Location, modalName: string, visible: bool, updatedState: Object) {
	if (location.state && location.state.modal && location.state.modal !== modalName) {
		console.error(`location.state.modal is not ${modalName}, is:`, location.state.modal);
	}

	router.replace({...location, state: {...location.state, modal: visible ? modalName : null, ...updatedState}});
}

// dismissModal creates a function that dismisses the modal by setting
// the location state's modal property to null.
export function dismissModal(modalName, location, router, updatedState) {
	return () => {
		setLocationModalState(router, location, modalName, false, updatedState);
	};
}

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, modalName, children, onDismiss}, {router}) {
	if (!location.state || location.state.modal !== modalName) return null;

	const onDismiss2 = () => {
		dismissModal(modalName, location, router)();
		if (onDismiss) onDismiss();
	};
	return (
		<Modal onDismiss={onDismiss2}>
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

	children: React.PropTypes.any,
};
LocationStateModal.contextTypes = {
	router: React.PropTypes.object.isRequired,
};
