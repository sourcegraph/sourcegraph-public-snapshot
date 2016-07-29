// @flow

import * as React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/menu.css";

class Menu extends React.Component {
	static propTypes = {
		children: React.PropTypes.any,
		className: React.PropTypes.string,
		style: React.PropTypes.object,
	};

	renderMenuItems() {
		return React.Children.map(this.props.children, function(ch) {
			return <div key={ch.props} styleName={`${ch.props.role ? ch.props.role : "inactive"}`}>{React.cloneElement(ch)}</div>;
		});
	}

	render() {
		return <div className={this.props.className} style={this.props.style} styleName="container">{this.renderMenuItems()}</div>;
	}
}


export default CSSModules(Menu, styles);
