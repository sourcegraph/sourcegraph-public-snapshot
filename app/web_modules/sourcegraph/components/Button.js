// @flow

import React from "react";

import Loader from "./Loader";
import CSSModules from "react-css-modules";
import styles from "./styles/button.css";

function Button({
	block = false,
	outline = false,
	size,
	disabled = false,
	loading = false,
	color = "default",
	onClick,
	imageUrl,
	children,
	...props,
}: {
	block: bool, // display:inline-block by default; use block for full-width buttons
	outline: bool, // solid by default
	size: "small" | "large",
	disabled: bool,
	loading: bool,
	color: "blue" | "purple" | "green" | "red" | "orange",
	onClick?: Function,
	imageUrl?: string,
	children: React$Element | Array<React$Element>,
}) {
	let style = `${outline ? "outline-" : "solid-"}${color}`;
	if (disabled || loading) style = `${style} disabled`;
	if (block) style = `${style} block`;
	style = `${style} ${size ? size : ""}`;

	return (
		<button {...props} styleName={style}
			onClick={onClick}>
			{imageUrl ? <img styleName="button-image" src={imageUrl} /> : ""}
			{loading && <Loader stretch={Boolean(block)} {...props}/>}
			{!loading && children}
		</button>
	);
}

export default CSSModules(Button, styles, {allowMultiple: true});
