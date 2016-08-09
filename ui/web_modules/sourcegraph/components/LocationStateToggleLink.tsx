// tslint:disable

import * as React from "react";

// Copied from react-router Link.js.
function isLeftClickEvent(ev) { return ev.button === 0; }
function isModifiedEvent(ev) { return Boolean(ev.metaKey || ev.altKey || ev.ctrlKey || ev.shiftKey); }

type Props = {
	location: any,

	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string,

	// href is the URL used if the user opens the link in
	// a new tab or copies the link.
	href?: string,

	// onToggle is called when the link is toggled ON.
	onToggle?: (v: boolean) => void,

	// target is the <a target=""> attribute value.
	target?: string,

	children?: any,

	role?: string,

	[key: string]: any,
};

// LocationStateToggleLink is like react-router's <Link>, but instead of going
// to a new URL, it merely toggles a boolean field on the location's state.
//
// It can be used for showing modals, whose on/off state should not be
// reflected in the URL. Something else will have to read the location state
// to determine whether to show it.
export function LocationStateToggleLink(props: Props, {router}) {
	const {location, children, modalName} = props;
	const other = Object.assign({}, props);
	delete other.location;
	delete other.children;
	delete other.modalName;
	const active = location.state && location.state.modal === modalName;

	// Copied from react-router Link.js.
	const handleClick = (ev) => {
		if (isModifiedEvent(ev) || !isLeftClickEvent(ev)) return;

		// If target prop is set (e.g., to "_blank"), let browser handle link.
		if (props.target) return;

		ev.preventDefault();
		router.push(Object.assign({}, location, {state: Object.assign({}, location.state, {modal: active ? null : modalName})}));

		if (props.onToggle) props.onToggle(!active);
	};

	return (
		<a {...other} // eslint-disable-line no-undef
			href={props.href}
			onClick={handleClick}>
			{children}
		</a>
	);
}

(LocationStateToggleLink as any).contextTypes = {
	router: React.PropTypes.object.isRequired,
};
