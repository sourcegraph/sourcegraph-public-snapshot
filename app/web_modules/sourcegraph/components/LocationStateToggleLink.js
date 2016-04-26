import React from "react";

// Copied from react-router Link.js.
function isLeftClickEvent(ev) { return ev.button === 0; }
function isModifiedEvent(ev) { return Boolean(ev.metaKey || ev.altKey || ev.ctrlKey || ev.shiftKey); }

// LocationStateToggleLink is like react-router's <Link>, but instead of going
// to a new URL, it merely toggles a boolean field on the location's state.
//
// It can be used for showing modals, whose on/off state should not be
// reflected in the URL. Something else will have to read the location state
// to determine whether to show it.
export default function LocationStateToggleLink(props, {router}) {
	const {location, children, stateKey, ...other} = props;
	const active = location.state && Boolean(location.state[stateKey]);

	// Copied from react-router Link.js.
	const handleClick = (ev) => {
		if (isModifiedEvent(ev) || !isLeftClickEvent(ev)) return;

		// If target prop is set (e.g., to "_blank"), let browser handle link.
		if (props.target) return;

		ev.preventDefault();
		router.push({...location, state: {...location.state, [stateKey]: !active}});
	};

	return (
		<a {...other} // eslint-disable-line no-undef
			href={props.href}
			onClick={handleClick}>
			{children}
		</a>
	);
}
LocationStateToggleLink.propTypes = {
	location: React.PropTypes.object.isRequired,

	// stateKey is the name of the key on the location's state that this
	// LocationStateToggleLink component toggles.
	stateKey: React.PropTypes.string.isRequired,

	// href is the URL used if the user opens the link in
	// a new tab or copies the link.
	href: React.PropTypes.string,
};
LocationStateToggleLink.contextTypes = {
	router: React.PropTypes.object.isRequired,
};
