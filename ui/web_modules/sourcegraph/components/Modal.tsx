// tslint:disable: typedef ordered-imports

import * as React from "react";

import * as styles from "sourcegraph/components/styles/modal.css";
import {renderedOnBody} from "sourcegraph/util/renderedOnBody";
import {EventListener} from "sourcegraph/Component";

interface ModalProps {
	onDismiss?: () => void;
}

type ModalState = any;

class ModalComp extends React.Component<ModalProps, ModalState> {
	constructor(props: ModalProps) {
		super(props);
		this._onClick = this._onClick.bind(this);
		this._handleKeydown = this._handleKeydown.bind(this);
	}

	componentDidMount() {
		// Prevent the page below the modal from scrolling.
		document.body.style.overflow = "hidden";
	}

	componentWillUnmount() {
		document.body.style.overflow = "";
	}

	_onClick(e) {
		if (e.target === this.refs["modal_container"]) {
			if (this.props.onDismiss) {
				this.props.onDismiss();
			}
		}
	}

	_handleKeydown(e: KeyboardEvent) {
		if (e.keyCode === 27 /* ESC */) {
			if (this.props.onDismiss) {
				this.props.onDismiss();
			}
		}
	}

	render(): JSX.Element | null {
		return (
			<div ref="modal_container"
					className={styles.container}
					onClick={this._onClick}>
				{this.props.children}
				<EventListener target={global.document} event="keydown" callback={this._handleKeydown} />
			</div>
		);
	}
}

let RenderedModal = renderedOnBody(ModalComp);

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

interface LocationStateModalProps {
	location: any;
	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string;
	onDismiss?: () => void;
	children?: any;
	style?: Object;
}

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, modalName, children, onDismiss, style}: LocationStateModalProps, {router}): any {
	if (!location.state || location.state.modal !== modalName) {
		return null;
	}

	const onDismiss2 = () => {
		dismissModal(modalName, location, router)();
		if (onDismiss) {
			onDismiss();
		}
	};

	return (
		<RenderedModal onDismiss={onDismiss2} style={style}>
			{children}
		</RenderedModal>
	);
}

(LocationStateModal as any).contextTypes = {
	router: React.PropTypes.object.isRequired,
};
