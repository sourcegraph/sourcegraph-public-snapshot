import * as React from "react";

import { Router } from "sourcegraph/app/router";

function isLeftClickEvent(ev: MouseEvent): boolean { return ev.button === 0; }
function isModifiedEvent(ev: MouseEvent): boolean { return Boolean(ev.metaKey || ev.altKey || ev.ctrlKey || ev.shiftKey); }

interface Props extends React.HTMLAttributes<HTMLAnchorElement> {
	location: any;
	// modalName is the name of the modal (location.state.modal value) that this
	// LocationStateToggleLink component toggles.
	modalName: string;

	// onToggle is called when the link is toggled ON.
	onToggle?: (v: boolean) => void;

	children?: any;
}

// LocationStateToggleLink is like react-router's <Link>, but instead of going
// to a new URL, it merely toggles a boolean field on the location's state.
//
// It can be used for showing modals, whose on/off state should not be
// reflected in the URL. Something else will have to read the location state
// to determine whether to show it.
export function LocationStateToggleLink(props: Props, { router }: { router: Router }): JSX.Element {
	let {
		children,
		modalName,
		onToggle,
		location,
		...rest,
	} = props;

	// Copied from react-router Link.js.
	const handleClick = (ev) => {
		if (isModifiedEvent(ev) || !isLeftClickEvent(ev)) {
			return;
		}

		// If target prop is set (e.g., to "_blank"), let browser handle link.
		if (props.target) {
			return;
		}

		ev.preventDefault();
		location = (router as any).getCurrentLocation();
		router.push({ ...location, state: { modal: modalName } });

		if (onToggle) {
			const state = location.state;
			const active = state && state.modal === modalName;
			onToggle(!active);
		}
	};

	return <a {...rest}
		onClick={handleClick}>
		{children}
	</a>;
}

(LocationStateToggleLink as any).contextTypes = {
	router: React.PropTypes.object.isRequired,
};
