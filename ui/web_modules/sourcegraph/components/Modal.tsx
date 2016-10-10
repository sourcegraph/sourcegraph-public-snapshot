import * as React from "react";
import {InjectedRouter} from "react-router";
import {EventListener} from "sourcegraph/Component";
import * as styles from "sourcegraph/components/styles/modal.css";
import {Location} from "sourcegraph/Location";
import {renderedOnBody} from "sourcegraph/util/renderedOnBody";

interface Props {
	onDismiss?: () => void;
	location?: Location;
}

interface State {
	originalOverflow: string | null;
}

export class ModalComp extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			originalOverflow: document.body.style.overflowY,
		};
		this._onClick = this._onClick.bind(this);
		this._handleKeydown = this._handleKeydown.bind(this);
	}

	componentDidMount(): void {
		this.setState({ originalOverflow: document.body.style.overflowY });
		document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		if (this.state.originalOverflow !== document.body.style.overflowY) {
			document.body.style.overflow = this.state.originalOverflow;
		}
	}

	_onClick(e: React.MouseEvent<HTMLElement>): void {
		if (e.target === this.refs["modal_container"]) {
			if (this.props.onDismiss) {
				this.props.onDismiss();
			}
		}
	}

	_handleKeydown(e: KeyboardEvent): void {
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
export function setLocationModalState(router: InjectedRouter, location: Location, modalName: string, visible: boolean): void {
	router.replace(Object.assign({},
		location,
		{
			state: Object.assign({},
				location.state,
				{ modal: visible ? modalName : null},
			),
		})
	);
}

// dismissModal creates a function that dismisses the modal by setting
// the location state's modal property to null.
export function dismissModal(modalName: string, location: Location, router: InjectedRouter): any {
	return () => {
		setLocationModalState(router, location, modalName, false);
	};
}

interface LocationStateModalProps {
	location: Location;
	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string;
	onDismiss?: (e: any) => void;
	children?: JSX.Element[];
	router: InjectedRouter;
	style?: React.CSSProperties;
}

// TODO(nicot): We are getting rid of this function below with the up and coming nicot modal refactor, so the casting I did below is temporary.
// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export function LocationStateModal({location, modalName, children, onDismiss, style, router}: LocationStateModalProps): JSX.Element {
	if (!location.state || !(location.state as any).modal || (location.state as any).modal  !== modalName) {
		return <span />;
	}

	const onDismiss2 = (e) => {
		dismissModal(modalName, location, router)();
		if (onDismiss) {
			onDismiss(e);
		}
	};

	return <RenderedModal onDismiss={onDismiss2} style={style} location={location} router={router}>
		{children}
	</RenderedModal>;
}
