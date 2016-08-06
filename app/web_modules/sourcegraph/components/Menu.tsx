// tslint:disable

import * as React from "react";

import CSSModules from "react-css-modules";
import * as styles from "./styles/menu.css";

class Menu extends React.Component<any, any> {
	static propTypes = {
		children: React.PropTypes.any,
		className: React.PropTypes.string,
		style: React.PropTypes.object,
	};

	renderMenuItems() {
		return React.Children.map(this.props.children, function(ch: React.ReactElement<any>) {
			return <div key={ch.props} styleName={`${ch.props.role ? ch.props.role : "inactive"}`}>{React.cloneElement(ch)}</div>;
		});
	}

	render(): JSX.Element | null {
		return <div className={`${this.props.className} ${styles.container}`} style={this.props.style}>{this.renderMenuItems()}</div>;
	}
}


export default CSSModules(Menu, styles);
