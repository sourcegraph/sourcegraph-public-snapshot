// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/panel.css";

class Panel extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		hoverLevel: React.PropTypes.string, // high, low
		hover: React.PropTypes.bool,
	};

	static defaultProps = {
		hoverLevel: "low",
		hover: false,
	};

	render() {
		const {children, hover, hoverLevel, className} = this.props;

		return (
			<div className={className}
				styleName={`panel ${hoverLevel} ${hover ? `${hoverLevel}-hover hover` : ""}`
			}>
				{children}
			</div>
		);
	}
}

export default CSSModules(Panel, styles, {allowMultiple: true});
