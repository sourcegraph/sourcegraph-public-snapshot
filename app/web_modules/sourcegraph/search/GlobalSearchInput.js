// @flow weak

import React from "react";
import {Input} from "sourcegraph/components";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearchInput.css";

function GlobalSearchInput(props) {
	// Omit styles prop so we don't clobber Input's own style mapping.
	const passProps = {...props, styles: undefined}; // eslint-disable-line no-undefined

	const cls = `${props.styles["search-input"]} ${props.styles[`${props.size}-input`]}`; // eslint-disable-line react/prop-types

	return (
		<Input type="text"
			block={true}
			spellCheck="off"
			autoComplete="off"
			className={cls}
			{...passProps} />
	);
}
GlobalSearchInput.propTypes = {
	size: React.PropTypes.oneOf(["small", "large"]).isRequired, // "large" (for homepage) or "small" (for global nav)
};

export default CSSModules(GlobalSearchInput, styles, {allowMultiple: true});
