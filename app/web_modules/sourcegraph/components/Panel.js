// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/panel.css";

class Panel extends React.Component {
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

	render() {
		const {children, color, inverse, hover, hoverLevel, className} = this.props;

		const colorClass = color ? `color ${inverse ? "inverse-" : ""}${color}` : "no-color";

		return (
			<div className={className}
				styleName={`panel ${colorClass} ${hoverLevel || ""} ${hover ? `${hoverLevel}-hover hover` : ""}`
			}>
				{children}
			</div>
		);
	}
}

export default CSSModules(Panel, styles, {allowMultiple: true});
