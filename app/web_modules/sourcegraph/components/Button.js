import * as React from "react";

import Loader from "./Loader";
import CSSModules from "react-css-modules";
import styles from "./styles/button.css";

function Button(props: {
	block: bool, // display:inline_block by default; use block for full-width buttons
	outline: bool, // solid by default
	size: "small" | "large",
	disabled: bool,
	loading: bool,
	color: "blue" | "purple" | "green" | "red" | "orange",
	onClick?: Function,
	imageUrl?: string,
	children: any,
}) {
	let {
		block = false,
		outline = false,
		size,
		disabled = false,
		loading = false,
		color = "normal",
		onClick,
		imageUrl,
		children,
	} = props;
	const other = Object.assign({}, props);
	delete other.block;
	delete other.outline;
	delete other.size;
	delete other.disabled;
	delete other.loading;
	delete other.color;
	delete other.onClick;
	delete other.imageUrl;
	delete other.children;

	let style = `${outline ? "outline_" : "solid_"}${color}`;
	if (disabled || loading) style = `${style} disabled`;
	if (block) style = `${style} block`;
	style = `${style} ${size ? size : ""}`;

	return (
		<button {...other} styleName={style}
			onClick={onClick}>
			{imageUrl ? <img styleName="button_image" src={imageUrl} /> : ""}
			{loading && <Loader stretch={Boolean(block)} {...props}/>}
			{!loading && children}
		</button>
	);
}

export default CSSModules(Button, styles, {allowMultiple: true});
