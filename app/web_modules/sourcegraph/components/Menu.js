// @flow weak

import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/menu.css";

function Menu(props, context) {
	return (
		<ol styleName="container">
			{React.Children.map(props.children, (ch) => (
				<li key={ch} styleName="item">{React.cloneElement(ch, {styleName: "item-content"})}</li>
			))}
		</ol>
	);
}
Menu.propTypes = {
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
	]),
};

export default CSSModules(Menu, styles);
