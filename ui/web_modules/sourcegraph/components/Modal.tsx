// tslint:disable: typedef ordered-imports curly

import * as React from "react";

import * as styles from "./styles/modal.css";
import {renderedOnBody} from "sourcegraph/util/renderedOnBody";

type ModalProps = {
	onDismiss?: () => void,
};

class ModalComp extends React.Component<ModalProps, any> {
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
		if (e.target === this.refs["modal_container"]) {
			if (this.props.onDismiss) this.props.onDismiss();
		}
	}

	_handleKeydown(e: KeyboardEvent) {
		if (e.keyCode === 27 /* ESC */) {
			if (this.props.onDismiss) this.props.onDismiss();
		}
	}

	render(): JSX.Element | null {
		return (
			<div ref="modal_container"
				className={styles.container}
				onClick={this._onClick}>
					{this.props.children}
			</div>
		);
	}
}

export const Modal = renderedOnBody(ModalComp);

// setLocationModalState shows or hides a modal by setting the location.state.modal
// property to modalName if shown is true and null otherwise.
export function setLocationModalState(router: any, location: any, modalName: string, visible: boolean) {
	if (location.state && location.state.modal && location.state.modal !== modalName) {
		console.error(`location.state.modal is not ${modalName}, is:`, location.state.modal);
	}

	router.replace(Object.assign({}, location, {state: Object.assign({}, location.state, {modal: visible ? modalName : null})}));
}

// dismissModal creates a function that dismisses the modal by setting
// the location state's modal property to null.
export function dismissModal(modalName, location, router) {
	return () => {
		setLocationModalState(router, location, modalName, false);
	};
}

type LocationStateModalProps = {
	location: any,

	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string,

	onDismiss?: () => void,

	children?: any,
};

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, modalName, children, onDismiss}: LocationStateModalProps, {router}): any {
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

(LocationStateModal as any).contextTypes = {
	router: React.PropTypes.object.isRequired,
};
