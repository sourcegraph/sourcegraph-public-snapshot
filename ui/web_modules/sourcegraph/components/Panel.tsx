// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "./styles/panel.css";
import * as classNames from "classnames";

type Props = {
	className?: string,
	children?: any,
	color?: string, // blue, white, purple, green, orange, (empty)
	inverse?: boolean, // light text on color background
	hoverLevel?: string, // high, low, (empty)
	hover?: boolean,
};

export class Panel extends React.Component<Props, any> {
	static defaultProps = {
		hover: false,
	};

	render(): JSX.Element | null {
		const {children, color, inverse, hover, hoverLevel, className} = this.props;

		return (
			<div {...this.props} className={classNames(styles.panel, colorClass(color || "", inverse || false), hoverClass(hoverLevel || "", hover || false), className)}>
				{children}
			</div>
		);
	}
}

function colorClass(color: string, inverse: boolean): string {
	switch (color) {
	case "blue":
		return `${styles.color} ${inverse ? styles.inverse_blue : styles.blue}`;
	case "white":
		return `${styles.color} ${inverse ? styles.inverse_white : styles.white}`;
	case "purple":
		return `${styles.color} ${inverse ? styles.inverse_purple : styles.purple}`;
	case "green":
		return `${styles.color} ${inverse ? styles.inverse_green : styles.green}`;
	case "orange":
		return `${styles.color} ${inverse ? styles.inverse_orange : styles.orange}`;
	default:
		return styles.no_color;
	}
}

function hoverClass(hoverLevel: string, hover: boolean): string {
	switch (hoverLevel) {
	case "high":
		if (hover) {
			return `${styles.high} ${styles.high_hover} ${styles.hover}`;
		}
		return `${styles.high}`;
	case "low":
		if (hover) {
			return `${styles.low} ${styles.low_hover} ${styles.hover}`;
		}
		return `${styles.low}`;
	default:
		return "";
	}
}
