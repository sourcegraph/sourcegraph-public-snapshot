import { hover, media } from "glamor";
import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { EventListener } from "sourcegraph/Component";
import { FlexContainer, Heading, Panel } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { getModalDismissedEventObject } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { renderedOnBody } from "sourcegraph/util/renderedOnBody";

interface Props {
	onDismiss?: () => void;
	location?: RouterLocation;
	children?: React.ReactNode[];
}

interface State {
	originalOverflow: string | null;
}

export class ModalComp extends React.Component<Props, State> {
	private htmlElement: HTMLElement;

	constructor(props: Props) {
		super(props);
		this.state = {
			originalOverflow: document.body.style.overflowY,
		};
		this._handleKeydown = this._handleKeydown.bind(this);
		this.bindBackingInstance = this.bindBackingInstance.bind(this);
	}

	componentDidMount(): void {
		this.setState({ originalOverflow: document.body.style.overflowY });
		document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		document.body.style.overflow = this.state.originalOverflow;
	}

	_handleKeydown(e: KeyboardEvent): void {
		if (e.keyCode === 27 /* ESC */) {
			if (this.props.onDismiss) {
				this.props.onDismiss();
			}
		}
	}

	bindBackingInstance(el: HTMLElement): void {
		this.htmlElement = el;
	}

	render(): JSX.Element | null {
		return <div ref={this.bindBackingInstance}>
			{this.props.children}
			<EventListener target={global.document} event="keydown" callback={this._handleKeydown} />
		</div>;
	}
}

const RenderedModal = renderedOnBody(ModalComp);

// setLocationModalState shows or hides a modal by setting the location.state.modal
// property to modalName if shown is true and null otherwise.
export function setLocationModalState(router: Router, modalName: string, visible: boolean): void {
	const location: RouterLocation = (router as any).getCurrentLocation();
	router.replace(Object.assign({},
		location,
		{
			state: Object.assign({},
				location.state,
				{ modal: visible ? modalName : null },
			),
			query: Object.assign({},
				location.query,
				{ modal: visible ? modalName : undefined },
			),
		})
	);
}

/**
 * Returns a function that dismisses the modal by unsetting the query or state
 * property in the location.
 */
export function dismissModal(modalName: string, router: Router): () => void {
	return () => {
		// Log all modal dismissal events in a consistent way. Note that any additions of new "modalName"s will require new events to be created
		const eventObject = getModalDismissedEventObject(modalName);
		if (eventObject) {
			eventObject.logEvent();
		} else {
			// TODO(dan) ensure proper params
		}

		setLocationModalState(router, modalName, false);
	};
}

interface LocationStateModalProps {
	// modalName is the name of the modal (location.{state,query}.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string;
	onDismiss?: (e: any) => void;
	children?: JSX.Element[];
	padded?: boolean;
	style?: React.CSSProperties;
	sticky?: boolean;
	title: string;
}

// LocationStateModal wraps <Modal> and uses a key on the location state
// to determine whether it is displayed. Use LocationStateModal with
// LocationStateToggleLink.
export class LocationStateModal extends React.Component<LocationStateModalProps, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
		const { modalName, children, padded, style, title } = this.props;
		const location = this.context.router.location;
		const currentModal = (location.state && location.state["modal"]) ? location.state["modal"] : location.query["modal"];
		if (currentModal !== modalName) {
			return <span />;
		}

		const onDismiss = (e) => {
			if (this.props.sticky) {
				return;
			}

			dismissModal(modalName, this.context.router)();
			if (this.props.onDismiss) {
				this.props.onDismiss(e);
			}
		};

		return <RenderedModal onDismiss={onDismiss} location={location} router={this.context.router}>
			<Modal style={style} title={title} padded={padded}
				{...!this.props.sticky && { onDismiss: onDismiss }}>
				{children}
			</Modal>
		</RenderedModal>;
	}
}

interface ModalProps {
	children?: React.ReactNode;
	padded?: boolean;
	title: string;
	style?: React.CSSProperties;
	onDismiss?: ((e: any) => void);
}

function Modal({ children, onDismiss, padded = true, style, title }: ModalProps): JSX.Element {

	const overlaySx = {
		zIndex: 3,
		position: "fixed",
		width: "100%",
		height: "100%",
		left: 0,
		top: 0,
		backgroundColor: colors.blueGrayD2(0.75),
		overflow: "auto",
	};

	// Clickable area to dismiss modal
	const backgroundSx = {
		position: "absolute",
		width: "100%",
		height: "100%",
		zIndex: -1,
	};

	const closeSx = {
		color: colors.blueGrayL1(),
		cursor: "pointer",
		padding: whitespace[3],
	};

	const modalPanelSx = {
		margin: "auto",
		marginTop: whitespace[8],
		marginBottom: whitespace[8],
		maxWidth: "30rem",
		...style,
	};

	const noMarginMobile = media(layout.breakpoints.sm, { marginTop: "0 !important" }).toString();

	return <div style={overlaySx}>
		<div onClick={onDismiss} style={backgroundSx} />
		<Panel hoverLevel="low" className={noMarginMobile} style={modalPanelSx}>
			<FlexContainer justify="between" items="center" style={{ borderBottom: `1px solid ${colors.blueGrayL2(0.5)}` }}>

				<Heading level={6} compact={true} style={{
					paddingLeft: whitespace[2],
					margin: whitespace[3],
				}}>{title}</Heading>

				{onDismiss && <div onClick={onDismiss} style={closeSx}
					{...hover({ color: `${colors.blueGray()} !important` }) }>
					<Close width={24} style={{ top: 0 }} />
				</div>}

			</FlexContainer>
			<div style={{ padding: padded && whitespace[5] }}>{children}</div>
		</Panel>
	</div>;
}
