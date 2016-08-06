// tslint:disable

import * as React from "react";
import * as styles from "sourcegraph/components/styles/hero.css";
import CSSModules from "react-css-modules";

class Hero extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		pattern: React.PropTypes.string,
		color: React.PropTypes.string, // white, purple, blue, green, dark
		children: React.PropTypes.any,
	};

	render(): JSX.Element | null {
		const {color, pattern, children, className} = this.props;

		return (
			<div className={`${styles.hero} ${colorClasses[color] || styles.white} ${patternClasses[pattern] || ""} ${className}`}>
				{children}
			</div>
		);
	}
}

const colorClasses = {
	"transparent": styles.transparent,
	"white": styles.white,
	"purple": styles.purple,
	"blue": styles.blue,
	"dark": styles.dark,
	"green": styles.green,
};

const patternClasses = {
	"objects": styles.bg_img_objects,
	"objects_fade": styles.bg_img_objects_fade,
};

export default CSSModules(Hero, styles, {allowMultiple: true});
