import * as React from "react";
import {Input} from "sourcegraph/components";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearchInput.css";
import base from "sourcegraph/components/styles/_base.css";
import invariant from "invariant";
import {Search} from "sourcegraph/components/symbols";

// If the user clicks the magnifying glass icon, the cursor should be
// placed at the end of the text, not the beginning. Without this event
// handler, these clicks would place the cursor at the beginning.
function positionCursorAtEndIfIconClicked(ev: MouseEvent) {

	if (ev.button !== 0) return;

	const input = ev.target;
	invariant(input instanceof HTMLInputElement, "target is not <input>");

	// See if we clicked on the magnifying glass.
	const b = input.getBoundingClientRect();
	const x = ev.clientX - b.left;
	const y = ev.clientY - b.top;
	// See if we clicked on the upper-padding of the element. Usually this moves
	// the selector to the beginning of the input field which is undesierable.
	const pt = parseInt(window.getComputedStyle(input, null).getPropertyValue("padding-top"), 10);

	const indent = parseInt(window.getComputedStyle(input, null).getPropertyValue("text-indent"), 10);
	invariant(indent > 0, "couldn't find input text-indent");

	// Focus at cursor if click is beyond the icon's bounds (with some pixels of buffer).
	if (x > (indent + 3) && y >= pt) return;

	ev.preventDefault();
	input.setSelectionRange(input.value.length, input.value.length);
	input.focus();
}

function GlobalSearchInput(props) {
	// Omit styles prop so we don't clobber Input's own style mapping.
	const passProps = {...props, styleName: undefined, styles: undefined}; // eslint-disable-line no-undefined

	return (
		<div styleName="flex-fill relative" className={base.mr3}>
			{props.icon &&
				<Search width={16} style={{top: "11px", left: "10px"}} styleName="absolute cool-mid-gray-fill layer-btm" />
			}
			<Input
				{...passProps}
				id="e2etest-search-input"
				type="text"
				onMouseDown={props.icon ? positionCursorAtEndIfIconClicked : null}
				block={true}
				autoCorrect="off"
				autoCapitalize="off"
				spellCheck="off"
				autoComplete="off"
				defaultValue={props.query}
				className={props.className || ""}
				style={{textIndent: props.icon ? "18px" : "0px", backgroundColor: "transparent"}} />
		</div>
	);
}
GlobalSearchInput.propTypes = {
	query: React.PropTypes.string.isRequired,
	icon: React.PropTypes.bool, // whether to show a magnifying glass icon
	border: React.PropTypes.bool,
	className: React.PropTypes.string,
};

export default CSSModules(GlobalSearchInput, styles, {allowMultiple: true});
