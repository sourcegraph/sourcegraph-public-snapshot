// tslint:disable

import * as React from "react";
import * as classNames from "classnames";

import {Loader} from "./Loader";
import * as styles from "./styles/button.css";

export function Button(props: {
	block?: boolean, // display:inline_block by default; use block for full-width buttons
	outline?: boolean, // solid by default
	size?: string,
	disabled?: boolean,
	loading?: boolean,
	color?: string,
	onClick?: Function,
	imageUrl?: string,
	children?: any,
	className?: string,
	type?: string,
	formNoValidate?: boolean,
	id?: string,
	tabIndex?: string,
	style?: any,
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
		className,
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
	delete other.className;

	return (
		<button {...(other as any)} className={classNames(colorClass(color, outline), (disabled || loading) && styles.disabled, block && styles.block, size && sizeClasses[size], className)}
			onClick={onClick}>
			{imageUrl ? <img className={styles.button_image} src={imageUrl} /> : ""}
			{loading && <Loader {...props}/>}
			{!loading && children}
		</button>
	);
}

function colorClass(color: string, outline: boolean): string {
	switch (color) {
	case "normal":
		return outline ? styles.outline_normal : styles.solid_normal;
	case "disabled":
		return outline ? "" : styles.solid_disabled;
	case "blue":
		return outline ? styles.outline_blue : styles.solid_blue;
	case "purple":
		return outline ? styles.outline_purple : styles.solid_purple;
	case "green":
		return outline ? styles.outline_green : styles.solid_green;
	case "red":
		return outline ? styles.outline_red : styles.solid_red;
	case "orange":
		return outline ? styles.outline_orange : styles.solid_orange;
	case "white":
		return outline ? "" : styles.solid_white;
	default:
		return outline ? styles.outline_normal : styles.solid_normal;
	}
}

const sizeClasses = {
	"large": styles.large,
	"small": styles.small,
};
