// @flow

import React from "react";
import styles from "sourcegraph/components/styles/hero.css";
import CSSModules from "react-css-modules";

class Hero extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		pattern: React.PropTypes.string,
		color: React.PropTypes.string,
		children: React.PropTypes.any,
	};

	render() {
		const {color, pattern, children, className} = this.props;

		let styleName = "hero ";
		styleName += color ? color : "white";
		styleName += pattern ? ` bg-img-${pattern}` : "";

		return (
			<div className={className} styleName={styleName}>
				{children}
			</div>
		);
	}
}


export default CSSModules(Hero, styles, {allowMultiple: true});

