// @flow

import * as React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/menu.css";

class Menu extends React.Component {
	static propTypes = {
		children: React.PropTypes.any,
		className: React.PropTypes.string,
	};

	renderMenuItems() {
		return React.Children.map(this.props.children, function(ch) {
			return <div key={ch.props} styleName={ch.props.role === "menu-item" ? "item" : "inactive"}>{React.cloneElement(ch)}</div>;
		});
	}

	render() {
		return <div className={this.props.className} styleName="container">{this.renderMenuItems()}</div>;
	}
}


export default CSSModules(Menu, styles);
