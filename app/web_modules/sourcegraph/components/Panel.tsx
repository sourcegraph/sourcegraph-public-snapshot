// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/panel.css";

class Panel extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		color: React.PropTypes.string, // blue, white, purple, green, orange, (empty)
		inverse: React.PropTypes.bool, // light text on color background
		hoverLevel: React.PropTypes.string, // high, low, (empty)
		hover: React.PropTypes.bool,
	};

	static defaultProps = {
		hover: false,
	};

	render(): JSX.Element | null {
		const {children, color, inverse, hover, hoverLevel, className} = this.props;

		return (
			<div {...this.props} className={`${styles.panel} ${colorClass(color, inverse)} ${hoverClass(hoverLevel, hover)} ${className}`}>
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

export default CSSModules(Panel, styles, {allowMultiple: true});
