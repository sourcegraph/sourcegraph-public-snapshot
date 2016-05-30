// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/heading.css";

class Heading extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		level: React.PropTypes.string, //  1 is the largest
		underline: React.PropTypes.string, // blue, purple, white, orange, green
		color: React.PropTypes.string, // purple, blue, green, orange, cool-mid-gray
		align: React.PropTypes.string, // left, right, center
	};

	static defaultProps = {
		level: "3", //  1 is the largest
		underline: null,
		color: null,
		align: null,
	};

	render() {
		const {className, children, level, color, underline, align} = this.props;

		return (
			<div className={className} styleName={`h${level} ${color ? color : ""} ${align ? align : ""}`}>
				{children}<br />
				{underline && <hr styleName={`line l-${underline}`} />}
			</div>
		);
	}
}

export default CSSModules(Heading, styles, {allowMultiple: true});
