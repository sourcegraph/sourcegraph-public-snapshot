// tslint:disable

import * as React from "react";
import styles from "sourcegraph/components/styles/hero.css";
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

		let styleName = "hero ";
		styleName += color ? color : "white";
		styleName += pattern ? ` bg_img_${pattern}` : "";

		return (
			<div className={className} styleName={styleName}>
				{children}
			</div>
		);
	}
}


export default CSSModules(Hero, styles, {allowMultiple: true});
